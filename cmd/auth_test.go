// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cloudmanic/harbor-cli/client"
	"github.com/cloudmanic/harbor-cli/config"
)

func TestApplyToken(t *testing.T) {
	creds := &config.Credentials{}
	before := time.Now().UnixMilli()
	applyToken(creds, &client.TokenResponse{
		AccessToken:  "at_new",
		RefreshToken: "rt_new",
		TokenType:    "Bearer",
		Scope:        "notes sync",
		ExpiresIn:    3600,
	})
	if creds.AccessToken != "at_new" || creds.RefreshToken != "rt_new" {
		t.Errorf("tokens not applied: %+v", creds)
	}
	if creds.Scope != "notes sync" {
		t.Errorf("scope = %q", creds.Scope)
	}
	// ExpiresAt should be ~1h in the future.
	wantMin := before + 3590*1000
	if creds.ExpiresAt < wantMin {
		t.Errorf("ExpiresAt = %d, want >= %d", creds.ExpiresAt, wantMin)
	}
}

func TestApplyTokenKeepsRefreshWhenAbsent(t *testing.T) {
	creds := &config.Credentials{RefreshToken: "rt_old"}
	applyToken(creds, &client.TokenResponse{AccessToken: "at_2", ExpiresIn: 60})
	if creds.RefreshToken != "rt_old" {
		t.Errorf("refresh token should be preserved when response omits it; got %q", creds.RefreshToken)
	}
}

func TestLoginSummaryJSONOmitsSecrets(t *testing.T) {
	creds := &config.Credentials{Email: "you@example.com", DeviceID: "cli-1", APIURL: "http://x/api/v1"}
	tok := &client.TokenResponse{AccessToken: "at_secret", RefreshToken: "rt_secret", Scope: "notes", ExpiresIn: 3600}

	var m map[string]any
	_ = json.Unmarshal(loginSummaryJSON(creds, tok, false), &m)
	if _, ok := m["access_token"]; ok {
		t.Error("access_token must be omitted unless --show-token")
	}
	if m["email"] != "you@example.com" || m["scope"] != "notes" {
		t.Errorf("summary = %v", m)
	}

	var m2 map[string]any
	_ = json.Unmarshal(loginSummaryJSON(creds, tok, true), &m2)
	if m2["access_token"] != "at_secret" {
		t.Error("access_token should be present with --show-token")
	}
}

func TestWhoamiJSON(t *testing.T) {
	creds := &config.Credentials{Email: "you@example.com", Scope: "notes", DeviceID: "cli-1", DeviceName: "dev"}
	var m map[string]any
	_ = json.Unmarshal(whoamiJSON(creds, true, false), &m)
	if m["token_valid"] != true {
		t.Errorf("token_valid = %v", m["token_valid"])
	}
	if _, ok := m["access_token"]; ok {
		t.Error("access_token must be omitted unless --show-token")
	}
}

func TestMapLoginError(t *testing.T) {
	cases := map[string]string{
		"invalid_grant":    "incorrect email or password",
		"email_unverified": "not verified",
		"invalid_client":   "unknown OAuth client",
	}
	for code, wantSub := range cases {
		got := mapLoginError(&client.APIError{Code: code, Message: "x"})
		if !strings.Contains(got.Error(), wantSub) {
			t.Errorf("mapLoginError(%s) = %q, want substring %q", code, got.Error(), wantSub)
		}
	}
	// A non-API error passes through.
	plain := mapLoginError(errorString("boom"))
	if plain.Error() != "boom" {
		t.Errorf("plain passthrough = %q", plain.Error())
	}
}

func TestNewDeviceID(t *testing.T) {
	id := newDeviceID()
	if !strings.HasPrefix(id, "cli-") {
		t.Errorf("device id = %q, want cli- prefix", id)
	}
	if id == newDeviceID() {
		t.Error("device ids should be unique")
	}
}

// errorString is a trivial error type for passthrough testing.
type errorString string

func (e errorString) Error() string { return string(e) }
