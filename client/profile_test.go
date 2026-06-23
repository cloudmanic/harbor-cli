// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

func TestGetProfile(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"id":"u1"}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).GetProfile(); err != nil {
		t.Fatalf("GetProfile error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/profile" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

func TestUpdateProfileUsesPUT(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"id":"u1"}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).UpdateProfile(map[string]any{"name": "Jane"}); err != nil {
		t.Fatalf("UpdateProfile error: %v", err)
	}
	if rec.Method != "PUT" || rec.Path != "/profile" {
		t.Errorf("%s %s, want PUT /profile", rec.Method, rec.Path)
	}
}

func TestChangePassword(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"changed":true}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ChangePassword("old", "new"); err != nil {
		t.Fatalf("ChangePassword error: %v", err)
	}
	if rec.Path != "/profile/change-password" {
		t.Errorf("path = %s", rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["current_password"] != "old" || body["new_password"] != "new" {
		t.Errorf("body = %v", body)
	}
}

func TestAvatarAndConfirmEmail(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{}}`)
	defer srv.Close()
	c := testClient(srv.URL)

	if _, err := c.SetAvatar("abc123"); err != nil {
		t.Fatalf("SetAvatar: %v", err)
	}
	if rec.Path != "/profile/avatar" {
		t.Errorf("set-avatar path = %s", rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["hash"] != "abc123" {
		t.Errorf("avatar body = %v (want hash)", body)
	}

	if _, err := c.RemoveAvatar(); err != nil {
		t.Fatalf("RemoveAvatar: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/profile/avatar" {
		t.Errorf("remove-avatar = %s %s", rec.Method, rec.Path)
	}

	if _, err := c.ConfirmEmailChange("ec_1"); err != nil {
		t.Fatalf("ConfirmEmailChange: %v", err)
	}
	if rec.Path != "/profile/email/confirm" {
		t.Errorf("confirm-email path = %s", rec.Path)
	}
}
