// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"net/http"
)

// postJSONNoRefresh performs a JSON POST that never attempts a transparent
// refresh. The OAuth/auth-recovery endpoints must use this: they are public
// (no bearer) and a refresh during a token exchange would be nonsensical and
// could rotate the refresh token recursively.
func (c *Client) postJSONNoRefresh(path string, body any) ([]byte, error) {
	full, err := c.buildURL(path, nil)
	if err != nil {
		return nil, err
	}
	var raw []byte
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	return c.request(http.MethodPost, full, raw, "application/json", false)
}

// PasswordGrant exchanges email + password for an access + refresh token (the
// primary login). scope, deviceID, and deviceName are optional. It returns the
// raw response bytes (for --json display) alongside the parsed token.
func (c *Client) PasswordGrant(clientID, username, password, scope, deviceID, deviceName string) ([]byte, *TokenResponse, error) {
	body := map[string]string{
		"grant_type": "password",
		"client_id":  clientID,
		"username":   username,
		"password":   password,
	}
	if scope != "" {
		body["scope"] = scope
	}
	if deviceID != "" {
		body["device_id"] = deviceID
	}
	if deviceName != "" {
		body["device_name"] = deviceName
	}
	data, err := c.postJSONNoRefresh("/oauth/token", body)
	if err != nil {
		return nil, nil, err
	}
	tok, err := DecodeToken(data)
	return data, tok, err
}

// RefreshGrant rotates a single-use refresh token into a new access + refresh
// pair. scope, when set, may only narrow the existing grant. Presenting an
// already-rotated token returns invalid_grant and revokes the whole family.
func (c *Client) RefreshGrant(clientID, refreshToken, scope string) ([]byte, *TokenResponse, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     clientID,
		"refresh_token": refreshToken,
	}
	if scope != "" {
		body["scope"] = scope
	}
	data, err := c.postJSONNoRefresh("/oauth/token", body)
	if err != nil {
		return nil, nil, err
	}
	tok, err := DecodeToken(data)
	return data, tok, err
}

// Revoke revokes a token (RFC 7009 style). A refresh token revokes its whole
// family; an access token revokes just itself. Always succeeds server-side,
// even for an unknown token (no validity leak).
func (c *Client) Revoke(token, tokenTypeHint string) error {
	body := map[string]string{"token": token}
	if tokenTypeHint != "" {
		body["token_type_hint"] = tokenTypeHint
	}
	_, err := c.postJSONNoRefresh("/oauth/revoke", body)
	return err
}

// Logout revokes the current session server-side (or every session when
// allDevices is true). Requires a bearer token.
func (c *Client) Logout(allDevices bool) error {
	_, err := c.doPost("/auth/logout", map[string]bool{"all_devices": allDevices})
	return err
}

// VerifyEmail consumes an email-verification token. Public.
func (c *Client) VerifyEmail(token string) ([]byte, error) {
	return c.postJSONNoRefresh("/auth/verify-email", map[string]string{"token": token})
}

// ResendVerification requests a fresh verification email. When a bearer is
// present the server uses it; otherwise the email body is required (and the
// response is always "sent", to avoid account enumeration).
func (c *Client) ResendVerification(email string) ([]byte, error) {
	body := map[string]string{}
	if email != "" {
		body["email"] = email
	}
	// Uses the standard path so a present bearer is honored; no refresh needed.
	return c.postJSONNoRefresh("/auth/verify-email/resend", body)
}

// ForgotPassword starts a password reset. Always reports success (anti-
// enumeration). Public.
func (c *Client) ForgotPassword(email string) ([]byte, error) {
	return c.postJSONNoRefresh("/auth/password/forgot", map[string]string{"email": email})
}

// ResetPassword completes a password reset with a reset token and new
// password. Revokes all sessions on success. Public.
func (c *Client) ResetPassword(token, password string) ([]byte, error) {
	return c.postJSONNoRefresh("/auth/password/reset", map[string]string{
		"token":    token,
		"password": password,
	})
}
