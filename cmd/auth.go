// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudmanic/harbor-cli/client"
	"github.com/cloudmanic/harbor-cli/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// loginSummaryJSON builds the --json output for login: a curated session
// summary that omits the raw token secrets unless showToken is set.
func loginSummaryJSON(creds *config.Credentials, tok *client.TokenResponse, showToken bool) []byte {
	m := map[string]any{
		"email":      creds.Email,
		"scope":      tok.Scope,
		"api_url":    creds.BaseURL(),
		"token_type": tok.TokenType,
		"expires_at": creds.ExpiresAt,
		"expires_in": tok.ExpiresIn,
		"device_id":  creds.DeviceID,
	}
	if showToken {
		m["access_token"] = tok.AccessToken
		m["refresh_token"] = tok.RefreshToken
	}
	out, _ := json.Marshal(m)
	return out
}

// whoamiJSON builds the --json output for whoami: session status without
// secrets unless showToken is set.
func whoamiJSON(creds *config.Credentials, valid, showToken bool) []byte {
	m := map[string]any{
		"email":       creds.Email,
		"api_url":     creds.BaseURL(),
		"scope":       creds.Scope,
		"token_valid": valid,
		"expires_at":  creds.ExpiresAt,
		"device_id":   creds.DeviceID,
		"device_name": creds.DeviceName,
	}
	if showToken {
		m["access_token"] = creds.AccessToken
	}
	out, _ := json.Marshal(m)
	return out
}

// tokenRefreshSkew is how far ahead of expiry we proactively refresh, so a
// request never races token expiry mid-flight.
const tokenRefreshSkew = 60 * time.Second

// applyToken copies a fresh token response into the credentials, computing the
// epoch-ms expiry from the OAuth expires_in (seconds).
func applyToken(creds *config.Credentials, tok *client.TokenResponse) {
	creds.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		creds.RefreshToken = tok.RefreshToken
	}
	if tok.TokenType != "" {
		creds.TokenType = tok.TokenType
	}
	if tok.Scope != "" {
		creds.Scope = tok.Scope
	}
	creds.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second).UnixMilli()
}

// refreshAndPersist rotates the refresh token and atomically persists the new
// pair. Persisting immediately is critical: a single-use refresh token that is
// rotated but not saved logs the user out. Returns the new access token and
// ok=false if the refresh failed (e.g. a revoked family → re-login needed).
func refreshAndPersist(c *client.Client, creds *config.Credentials) (string, bool) {
	_, tok, err := c.RefreshGrant(creds.EffectiveClientID(), creds.RefreshToken, "")
	if err != nil {
		return "", false
	}
	applyToken(creds, tok)
	if err := config.Save(creds); err != nil {
		return "", false
	}
	return tok.AccessToken, true
}

// newDeviceID mints a stable per-install device id (used for session/device
// tracking and sync). Generated once at first login and reused thereafter.
func newDeviceID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "cli-unknown"
	}
	return "cli-" + hex.EncodeToString(b)
}

// defaultDeviceName describes this install for the session list.
func defaultDeviceName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown-host"
	}
	return "harbor-cli on " + host
}

// sharedStdin is a single buffered reader over os.Stdin, lazily created. Multiple
// non-interactive prompts in one invocation (e.g. crypto rotate's new + confirm)
// MUST read from the same reader — a fresh bufio.Reader per call would let the
// first read buffer past its newline and leave the next call at EOF.
var sharedStdin *bufio.Reader

// stdinReader returns the process-wide buffered stdin reader.
func stdinReader() *bufio.Reader {
	if sharedStdin == nil {
		sharedStdin = bufio.NewReader(os.Stdin)
	}
	return sharedStdin
}

// promptLine reads a single trimmed line from stdin after printing a prompt.
func promptLine(label string) (string, error) {
	fmt.Print(label)
	line, err := stdinReader().ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// promptPassword reads a password without echoing it. On an interactive
// terminal it prints the label and reads with echo disabled. When stdin is a
// pipe (scripts, CI, AI agents), it reads a single line from stdin instead, so
// `printf 'secret\n' | harbor login --email …` works non-interactively. The
// bytes are never stored or logged.
func promptPassword(label string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		line, err := stdinReader().ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
	fmt.Print(label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(b), nil
}

// ===========================================================================
// Commands
// ===========================================================================

// authCmd is the parent for the auth lifecycle helpers (refresh, status, and
// the public email/password recovery flows).
var authCmd = &cobra.Command{
	Use:     "auth",
	Short:   "Authentication helpers (refresh, status, email/password recovery)",
	GroupID: groupAuth,
}

// loginCmd performs the OAuth2 password grant and persists the session.
var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Log in with your email and password",
	GroupID: groupAuth,
	Long: `Log in to Harbor using the OAuth2 password grant.

You are prompted for your password (never echoed, never stored as plaintext);
only the resulting access + refresh tokens are saved to
~/.config/harbor/credentials.json (0600). Tokens refresh transparently.`,
	Example: `  harbor login
  harbor login --email you@example.com
  harbor login --scope "notes notebooks sync"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		scope, _ := cmd.Flags().GetString("scope")
		clientID, _ := cmd.Flags().GetString("client-id")
		showToken, _ := cmd.Flags().GetBool("show-token")

		var err error
		if email == "" {
			if email, err = promptLine("Email: "); err != nil {
				return err
			}
		}
		if email == "" {
			return errors.New("email is required")
		}
		password, err := promptPassword("Password: ")
		if err != nil {
			return err
		}
		if password == "" {
			return errors.New("password is required")
		}

		if clientID == "" {
			clientID = config.DefaultClientID
		}

		// Reuse an existing device id when one is already saved.
		deviceID := ""
		deviceName := defaultDeviceName()
		if existing, lerr := config.Load(); lerr == nil {
			deviceID = existing.DeviceID
			if existing.DeviceName != "" {
				deviceName = existing.DeviceName
			}
		}
		if deviceID == "" {
			deviceID = newDeviceID()
		}

		c := newAnonymousClient()
		_, tok, err := c.PasswordGrant(clientID, email, password, scope, deviceID, deviceName)
		if err != nil {
			return mapLoginError(err)
		}

		// Persist the session. Store api_url only when it differs from the
		// production default, so a normal login still honors HARBOR_API_URL.
		apiToStore := resolveBaseURL(nil)
		if apiToStore == config.DefaultBaseURL {
			apiToStore = ""
		}
		creds := &config.Credentials{
			APIURL:     apiToStore,
			ClientID:   clientID,
			Email:      email,
			DeviceID:   deviceID,
			DeviceName: deviceName,
		}
		applyToken(creds, tok)
		if err := config.Save(creds); err != nil {
			return err
		}

		if jsonOutput {
			printResult(loginSummaryJSON(creds, tok, showToken), func([]byte) {})
			return nil
		}
		fmt.Printf("Logged in as %s\n", bold(email))
		fmt.Printf("Scopes: %s\n", tok.Scope)
		fmt.Printf("Token expires %s (%s)\n", relTime(float64(creds.ExpiresAt)), epochMS(float64(creds.ExpiresAt)))
		if showToken {
			fmt.Printf("Access token: %s\n", tok.AccessToken)
		}
		return nil
	},
}

// logoutCmd revokes the session server-side and clears local credentials.
var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Log out and clear saved credentials",
	GroupID: groupAuth,
	Long:    "Revokes the current session on the server (or all sessions with --all-devices) and removes the local credentials file. Idempotent.",
	Example: `  harbor logout
  harbor logout --all-devices`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allDevices, _ := cmd.Flags().GetBool("all-devices")

		creds, err := config.Load()
		if err != nil {
			fmt.Println("Not logged in.")
			return nil
		}

		c, _, err := loadClientFromConfig()
		if err == nil {
			// Best-effort server-side revocation; clear locally regardless.
			_ = c.Logout(allDevices)
			if creds.RefreshToken != "" {
				_ = c.Revoke(creds.RefreshToken, "refresh_token")
			}
		}
		if err := config.Clear(); err != nil {
			return err
		}
		// Drop the cached (wrapped) keystore blob too, so a logout leaves no
		// account-specific encryption state behind.
		_ = config.ClearKeystoreBlob()
		if allDevices {
			fmt.Println("Logged out of all devices.")
		} else {
			fmt.Println("Logged out.")
		}
		return nil
	},
}

// whoamiCmd shows the current session status.
var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "Show the current session (alias: auth status)",
	GroupID: groupAuth,
	Long:    "Displays the logged-in email, API target, granted scopes, token expiry, and device identity.",
	RunE:    runWhoami,
}

// authStatusCmd is the `auth status` alias of whoami.
var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current session status",
	RunE:  runWhoami,
}

// runWhoami renders the saved session details (offline; no network call).
func runWhoami(cmd *cobra.Command, args []string) error {
	creds, err := config.Load()
	if err != nil {
		return err
	}
	showToken, _ := cmd.Flags().GetBool("show-token")
	valid := !creds.IsExpired(0)

	if jsonOutput {
		printResult(whoamiJSON(creds, valid, showToken), func([]byte) {})
		return nil
	}
	apiURL := creds.BaseURL()
	pairs := [][2]string{
		{"Email", creds.Email},
		{"API URL", apiURL},
		{"Scopes", creds.Scope},
		{"Token valid", boolMark(valid)},
		{"Expires", fmt.Sprintf("%s (%s)", relTime(float64(creds.ExpiresAt)), epochMS(float64(creds.ExpiresAt)))},
		{"Device", fmt.Sprintf("%s (%s)", creds.DeviceName, creds.DeviceID)},
	}
	if showToken {
		pairs = append(pairs, [2]string{"Access token", creds.AccessToken})
	}
	printKV(pairs)
	if !valid {
		fmt.Println(dim("Access token is expired; it will refresh on the next command."))
	}
	return nil
}

// authRefreshCmd forces an immediate token refresh.
var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Force a token refresh now",
	Long:  "Rotates the refresh token immediately and prints the new expiry. Tokens normally refresh transparently; this is for diagnostics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.Load()
		if err != nil {
			return err
		}
		if creds.RefreshToken == "" {
			return errors.New("no refresh token saved — run 'harbor login'")
		}
		c := client.NewClient(resolveBaseURL(creds), creds.AccessToken)
		c.Verbose = verboseFlag
		data, tok, err := c.RefreshGrant(creds.EffectiveClientID(), creds.RefreshToken, "")
		if err != nil {
			return mapRefreshError(err)
		}
		applyToken(creds, tok)
		if err := config.Save(creds); err != nil {
			return err
		}
		if jsonOutput {
			printResult(data, func([]byte) {})
			return nil
		}
		fmt.Printf("Token refreshed. Expires %s (%s)\n", relTime(float64(creds.ExpiresAt)), epochMS(float64(creds.ExpiresAt)))
		return nil
	},
}

// verifyEmailCmd consumes an email-verification token (public).
var verifyEmailCmd = &cobra.Command{
	Use:     "verify-email",
	Short:   "Verify your email with a verification token",
	Example: "  harbor auth verify-email --token ev_...",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			return errors.New("--token is required")
		}
		data, err := newAnonymousClient().VerifyEmail(token)
		if err != nil {
			return err
		}
		printResult(data, displayMessage("Email verified."))
		return nil
	},
}

// resendVerificationCmd requests a new verification email (anti-enumeration).
var resendVerificationCmd = &cobra.Command{
	Use:   "resend-verification",
	Short: "Resend the email-verification message",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		// Prefer an authenticated call when logged in; otherwise send the email.
		c := newAnonymousClient()
		if creds, err := config.Load(); err == nil && creds.AccessToken != "" {
			c = client.NewClient(resolveBaseURL(creds), creds.AccessToken)
		} else if email == "" {
			if email, err = promptLine("Email: "); err != nil {
				return err
			}
		}
		data, err := c.ResendVerification(email)
		if err != nil {
			return err
		}
		printResult(data, displayMessage("If that email exists and is unverified, a verification message was sent."))
		return nil
	},
}

// forgotPasswordCmd starts a password reset (anti-enumeration).
var forgotPasswordCmd = &cobra.Command{
	Use:     "forgot-password",
	Short:   "Request a password-reset email",
	Example: "  harbor auth forgot-password --email you@example.com",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		var err error
		if email == "" {
			if email, err = promptLine("Email: "); err != nil {
				return err
			}
		}
		data, err := newAnonymousClient().ForgotPassword(email)
		if err != nil {
			return err
		}
		printResult(data, displayMessage("If that email exists, a password-reset message was sent."))
		return nil
	},
}

// resetPasswordCmd completes a password reset with a token + new password.
var resetPasswordCmd = &cobra.Command{
	Use:     "reset-password",
	Short:   "Reset your password using a reset token",
	Long:    "Completes a password reset. Prompts for the new password (hidden). All existing sessions are revoked on success.",
	Example: "  harbor auth reset-password --token pr_...",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			return errors.New("--token is required")
		}
		pw, err := promptPassword("New password: ")
		if err != nil {
			return err
		}
		confirm, err := promptPassword("Confirm new password: ")
		if err != nil {
			return err
		}
		if pw != confirm {
			return errors.New("passwords do not match")
		}
		data, err := newAnonymousClient().ResetPassword(token, pw)
		if err != nil {
			return mapWeakPassword(err)
		}
		printResult(data, displayMessage("Password reset. All sessions were revoked — run 'harbor login' to sign in again."))
		return nil
	},
}

// mapLoginError translates raw password-grant errors into friendly guidance
// without leaking which factor was wrong (no account enumeration).
func mapLoginError(err error) error {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case "invalid_grant":
			return errors.New("incorrect email or password")
		case "email_unverified":
			return errors.New("your email is not verified — check your inbox, or run 'harbor auth resend-verification'")
		case "invalid_client":
			return errors.New("unknown OAuth client — check --client-id")
		case "invalid_scope":
			return errors.New("requested scope exceeds what this client allows")
		}
	}
	return err
}

// mapRefreshError gives a clear "session expired" message on a dead refresh
// token (e.g. a revoked family).
func mapRefreshError(err error) error {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.Code == "invalid_grant" {
		return errors.New("session expired — run 'harbor login'")
	}
	return err
}

// mapWeakPassword surfaces the strength-policy details on a rejected password.
func mapWeakPassword(err error) error {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && (apiErr.Code == "weak_password" || apiErr.Code == "password_reused") {
		return err // renderError already prints details bullets
	}
	return err
}

func init() {
	loginCmd.Flags().String("email", "", "Account email (prompted if omitted)")
	loginCmd.Flags().String("scope", "", "Space-delimited scopes to request (default: the client's full set)")
	loginCmd.Flags().String("client-id", "", "OAuth client id (default: harbor-app)")
	loginCmd.Flags().Bool("show-token", false, "Print the access token in the output")

	logoutCmd.Flags().Bool("all-devices", false, "Revoke every session, not just this one")

	whoamiCmd.Flags().Bool("show-token", false, "Include the access token in the output")
	authStatusCmd.Flags().Bool("show-token", false, "Include the access token in the output")

	verifyEmailCmd.Flags().String("token", "", "Email-verification token (required)")
	resendVerificationCmd.Flags().String("email", "", "Email to send to (when not logged in)")
	forgotPasswordCmd.Flags().String("email", "", "Account email")
	resetPasswordCmd.Flags().String("token", "", "Password-reset token (required)")

	authCmd.AddCommand(authRefreshCmd, authStatusCmd, verifyEmailCmd, resendVerificationCmd, forgotPasswordCmd, resetPasswordCmd)
	rootCmd.AddCommand(loginCmd, logoutCmd, whoamiCmd, authCmd)
}
