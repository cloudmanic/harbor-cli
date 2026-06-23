// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

func TestPasswordGrant(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"access_token":"at_1","refresh_token":"rt_1","token_type":"Bearer","expires_in":3600,"scope":"notes"}`)
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, tok, err := c.PasswordGrant("harbor-app", "you@example.com", "pw", "notes", "dev-1", "Dev")
	if err != nil {
		t.Fatalf("PasswordGrant error: %v", err)
	}
	if tok.AccessToken != "at_1" {
		t.Errorf("token = %+v", tok)
	}
	if rec.Path != "/oauth/token" {
		t.Errorf("path = %s", rec.Path)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body, &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["grant_type"] != "password" || body["username"] != "you@example.com" {
		t.Errorf("body = %v", body)
	}
	if body["device_id"] != "dev-1" {
		t.Errorf("device_id not sent: %v", body)
	}
}

func TestRefreshGrant(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"access_token":"at_2","refresh_token":"rt_2","token_type":"Bearer","expires_in":3600,"scope":"notes"}`)
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, tok, err := c.RefreshGrant("harbor-app", "rt_1", "")
	if err != nil {
		t.Fatalf("RefreshGrant error: %v", err)
	}
	if tok.RefreshToken != "rt_2" {
		t.Errorf("rotated token = %+v", tok)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["grant_type"] != "refresh_token" || body["refresh_token"] != "rt_1" {
		t.Errorf("body = %v", body)
	}
}

// TestRefreshGrantNeverRecurses ensures the token endpoint path does not invoke
// a refresh hook even on a 401 (it must use the no-refresh request path).
func TestRefreshGrantNeverRecurses(t *testing.T) {
	srv := newTestServer(t, nil, 401, `{"error":{"code":"invalid_grant","message":"bad"}}`)
	defer srv.Close()
	c := NewClient(srv.URL, "")
	hookCalls := 0
	c.OnUnauthorized = func() (string, bool) { hookCalls++; return "x", true }
	_, _, err := c.RefreshGrant("harbor-app", "rt_dead", "")
	if err == nil {
		t.Fatal("expected invalid_grant error")
	}
	if hookCalls != 0 {
		t.Errorf("refresh hook invoked %d times during a token call; want 0", hookCalls)
	}
}

func TestLogoutAndRevoke(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	defer srv.Close()
	c := testClient(srv.URL)
	if err := c.Logout(true); err != nil {
		t.Fatalf("Logout error: %v", err)
	}
	if rec.Path != "/auth/logout" {
		t.Errorf("path = %s", rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["all_devices"] != true {
		t.Errorf("all_devices not sent: %v", body)
	}

	if err := c.Revoke("rt_1", "refresh_token"); err != nil {
		t.Fatalf("Revoke error: %v", err)
	}
	if rec.Path != "/oauth/revoke" {
		t.Errorf("revoke path = %s", rec.Path)
	}
}

func TestPublicAuthHelpers(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(c *Client) error
		path string
	}{
		{"verify-email", func(c *Client) error { _, e := c.VerifyEmail("ev_1"); return e }, "/auth/verify-email"},
		{"resend", func(c *Client) error { _, e := c.ResendVerification("you@example.com"); return e }, "/auth/verify-email/resend"},
		{"forgot", func(c *Client) error { _, e := c.ForgotPassword("you@example.com"); return e }, "/auth/password/forgot"},
		{"reset", func(c *Client) error { _, e := c.ResetPassword("pr_1", "newpw"); return e }, "/auth/password/reset"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rec recordedRequest
			srv := newTestServer(t, &rec, 200, `{"data":{"ok":true}}`)
			defer srv.Close()
			if err := tc.call(NewClient(srv.URL, "")); err != nil {
				t.Fatalf("%s error: %v", tc.name, err)
			}
			if rec.Path != tc.path {
				t.Errorf("path = %s, want %s", rec.Path, tc.path)
			}
		})
	}
}
