// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

// TestAccountDeleteGuard exercises the non-interactive / confirmation-phrase
// guard, which is the security-critical bit of the destructive delete flow.
func TestAccountDeleteGuard(t *testing.T) {
	const phrase = accountDeleteConfirmPhrase

	cases := []struct {
		name        string
		jsonMode    bool
		interactive bool
		confirm     string
		yes         bool
		wantPhrase  string // expected returned phrase ("" = defer to prompt)
		wantErr     string // substring expected in the error ("" = no error)
	}{
		// Non-interactive: both --confirm (verbatim) and --yes required.
		{"noninteractive ok", false, false, phrase, true, phrase, ""},
		{"json ok", true, true, phrase, true, phrase, ""},
		{"noninteractive missing yes", false, false, phrase, false, "", "--yes"},
		{"noninteractive missing confirm", false, false, "", true, "", "--confirm"},
		{"noninteractive wrong confirm", false, false, "delete my account", true, "", "--confirm"},
		{"json without yes", true, true, phrase, false, "", "--yes"},

		// Interactive: empty confirm defers to the prompt; a wrong one fails fast;
		// a correct one is accepted.
		{"interactive defer", false, true, "", false, "", ""},
		{"interactive presupplied ok", false, true, phrase, false, phrase, ""},
		{"interactive wrong confirm", false, true, "nope", false, "", "did not match"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := accountDeleteGuard(tc.jsonMode, tc.interactive, tc.confirm, tc.yes)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want substring %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantPhrase {
				t.Errorf("phrase = %q, want %q", got, tc.wantPhrase)
			}
		})
	}
}

// TestAccountDeleteGuardRejectsCaseFold confirms the phrase is matched verbatim,
// not case-folded or trimmed.
func TestAccountDeleteGuardRejectsCaseFold(t *testing.T) {
	for _, bad := range []string{"delete my account", " DELETE MY ACCOUNT", "DELETE MY ACCOUNT "} {
		if _, err := accountDeleteGuard(true, true, bad, true); err == nil {
			t.Errorf("guard accepted non-verbatim phrase %q", bad)
		}
	}
}

// TestMapAccountError maps the domain codes to friendly messages.
func TestMapAccountError(t *testing.T) {
	cases := map[string]string{
		"confirmation_mismatch": "did not match",
		"already_scheduled":     "already pending",
		"not_scheduled":         "no account deletion is pending",
		"grace_expired":         "window has passed",
		"reauth_required":       "incorrect current password",
		"not_found":             "no such export job",
	}
	for code, sub := range cases {
		got := mapAccountError(apiErr(code))
		if !strings.Contains(got.Error(), sub) {
			t.Errorf("mapAccountError(%s) = %q, want substring %q", code, got.Error(), sub)
		}
	}
}

// TestDisplayExportJobQueued renders a freshly-started job (export_job_id form).
func TestDisplayExportJobQueued(t *testing.T) {
	data := []byte(`{"data":{"export_job_id":"e1","status":"queued"}}`)
	out := captureStdout(t, func() { displayExportJob(data) })
	if !strings.Contains(out, "e1") || !strings.Contains(out, "queued") {
		t.Errorf("queued job view missing fields:\n%s", out)
	}
}

// TestDisplayExportJobCompleted renders a completed job with progress and URL.
func TestDisplayExportJobCompleted(t *testing.T) {
	data := []byte(`{"data":{"id":"e1","status":"completed","total_units":4,"done_units":4,"download_url":"https://s3/x?sig=1","result_expires_at":1750003600000}}`)
	out := captureStdout(t, func() { displayExportJob(data) })
	if !strings.Contains(out, "4/4 notebooks") {
		t.Errorf("progress missing:\n%s", out)
	}
	if !strings.Contains(out, "https://s3/x") {
		t.Errorf("download URL missing:\n%s", out)
	}
	if !strings.Contains(out, "--download") {
		t.Errorf("download hint missing:\n%s", out)
	}
}

// TestDisplayDeletionScheduled surfaces the purge window and the cancel hint.
func TestDisplayDeletionScheduled(t *testing.T) {
	data := []byte(`{"data":{"status":"scheduled","purge_after":1752592000000,"grace_days":30,"can_cancel_until":1752592000000}}`)
	out := captureStdout(t, func() { displayDeletionScheduled(data) })
	if !strings.Contains(out, "scheduled") || !strings.Contains(out, "30") {
		t.Errorf("scheduled view missing fields:\n%s", out)
	}
	if !strings.Contains(out, "cancel-delete") {
		t.Errorf("cancel hint missing:\n%s", out)
	}
}
