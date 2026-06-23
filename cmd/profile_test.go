// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

func TestDisplayProfile(t *testing.T) {
	data := []byte(`{"data":{"id":"u1","name":"Jane","email":"jane@example.com","email_verified":true,"pending_email":"new@example.com","locale":"en","timezone":"UTC","created_at":1750000000000,"updated_at":1750000000000}}`)
	out := captureStdout(t, func() { displayProfile(data) })
	if !strings.Contains(out, "jane@example.com") {
		t.Errorf("email missing:\n%s", out)
	}
	if !strings.Contains(out, "Pending email") || !strings.Contains(out, "new@example.com") {
		t.Errorf("pending email missing:\n%s", out)
	}
}

func TestMapProfileError(t *testing.T) {
	cases := map[string]string{
		"reauth_required": "current password",
		"email_taken":     "already in use",
		"password_reused": "differ",
	}
	for code, sub := range cases {
		if got := mapProfileError(apiErr(code)); !strings.Contains(got.Error(), sub) {
			t.Errorf("mapProfileError(%s) = %q", code, got.Error())
		}
	}
}
