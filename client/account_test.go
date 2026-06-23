// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"strings"
	"testing"
)

// TestStartAccountExport verifies the export start posts to /account/export.
func TestStartAccountExport(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 202, `{"data":{"export_job_id":"e1","status":"queued"}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).StartAccountExport(); err != nil {
		t.Fatalf("StartAccountExport error: %v", err)
	}
	if rec.Method != "POST" {
		t.Errorf("method = %s", rec.Method)
	}
	if rec.Path != "/account/export" {
		t.Errorf("path = %s", rec.Path)
	}
	if rec.Auth != "Bearer at_test_token" {
		t.Errorf("auth = %q", rec.Auth)
	}
}

// TestGetAccountExport verifies the poll hits GET /account/export/:id.
func TestGetAccountExport(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"id":"e1","status":"completed","download_url":"https://s3/x"}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).GetAccountExport("e1"); err != nil {
		t.Fatalf("GetAccountExport error: %v", err)
	}
	if rec.Method != "GET" {
		t.Errorf("method = %s", rec.Method)
	}
	if rec.Path != "/account/export/e1" {
		t.Errorf("path = %s", rec.Path)
	}
}

// TestRequestAccountDeletion verifies the delete request posts both the password
// (as current_password) and the confirmation phrase (as confirm).
func TestRequestAccountDeletion(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"status":"scheduled","purge_after":1752592000000,"grace_days":30}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).RequestAccountDeletion("hunter2", "DELETE MY ACCOUNT"); err != nil {
		t.Fatalf("RequestAccountDeletion error: %v", err)
	}
	if rec.Method != "POST" {
		t.Errorf("method = %s", rec.Method)
	}
	if rec.Path != "/account/delete" {
		t.Errorf("path = %s", rec.Path)
	}
	body := string(rec.Body)
	if !containsAll(body, `"current_password"`, "hunter2", `"confirm"`, "DELETE MY ACCOUNT") {
		t.Errorf("body missing fields: %s", body)
	}
}

// TestCancelAccountDeletion verifies cancel posts the password to the cancel path.
func TestCancelAccountDeletion(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"status":"active"}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).CancelAccountDeletion("hunter2"); err != nil {
		t.Fatalf("CancelAccountDeletion error: %v", err)
	}
	if rec.Method != "POST" {
		t.Errorf("method = %s", rec.Method)
	}
	if rec.Path != "/account/delete/cancel" {
		t.Errorf("path = %s", rec.Path)
	}
	body := string(rec.Body)
	if !strings.Contains(body, "hunter2") || strings.Contains(body, "confirm") {
		t.Errorf("cancel body should carry only current_password: %s", body)
	}
}
