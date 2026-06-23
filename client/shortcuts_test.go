// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

// TestListShortcuts verifies the list method hits GET /shortcuts and forwards
// the standard list params as query string.
func TestListShortcuts(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ListShortcuts(map[string]string{"order": "-usn", "include_deleted": "true"}); err != nil {
		t.Fatalf("ListShortcuts error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/shortcuts" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if rec.Query == "" || !containsAll(rec.Query, "order=-usn", "include_deleted=true") {
		t.Errorf("query = %q", rec.Query)
	}
}

// TestGetShortcutIncludeDeleted verifies the get method targets the id path and
// sends include_deleted when requested.
func TestGetShortcutIncludeDeleted(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"id":"s1"}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).GetShortcut("s1", true); err != nil {
		t.Fatalf("GetShortcut error: %v", err)
	}
	if rec.Path != "/shortcuts/s1" {
		t.Errorf("path = %s", rec.Path)
	}
	if rec.Query != "include_deleted=true" {
		t.Errorf("query = %q", rec.Query)
	}
}

// TestCreateShortcut verifies a record shortcut is POSTed with its type and
// target_id in the JSON body.
func TestCreateShortcut(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"id":"s1","type":"note","target_id":"n1"}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).CreateShortcut(map[string]any{"type": "note", "target_id": "n1", "label": "Plan"}); err != nil {
		t.Fatalf("CreateShortcut error: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/shortcuts" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["type"] != "note" || body["target_id"] != "n1" || body["label"] != "Plan" {
		t.Errorf("body = %v", body)
	}
}

// TestUpdateShortcut verifies a partial update is PATCHed to the id path.
func TestUpdateShortcut(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"id":"s1","label":"New"}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).UpdateShortcut("s1", map[string]any{"label": "New"}); err != nil {
		t.Fatalf("UpdateShortcut error: %v", err)
	}
	if rec.Method != "PATCH" || rec.Path != "/shortcuts/s1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["label"] != "New" {
		t.Errorf("body = %v", body)
	}
}

// TestDeleteShortcut verifies the delete method issues DELETE to the id path.
func TestDeleteShortcut(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	defer srv.Close()
	if _, err := testClient(srv.URL).DeleteShortcut("s1"); err != nil {
		t.Fatalf("DeleteShortcut error: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/shortcuts/s1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

// TestReorderShortcuts verifies the bulk reorder PUTs the complete ordered id
// list to the literal /shortcuts/order route.
func TestReorderShortcuts(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ReorderShortcuts([]string{"s1", "s2", "s3"}); err != nil {
		t.Fatalf("ReorderShortcuts error: %v", err)
	}
	if rec.Method != "PUT" || rec.Path != "/shortcuts/order" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body struct {
		Order []string `json:"order"`
	}
	_ = json.Unmarshal(rec.Body, &body)
	if len(body.Order) != 3 || body.Order[0] != "s1" || body.Order[2] != "s3" {
		t.Errorf("order body = %v", body.Order)
	}
}

// TestReorderShortcutsNilEncodesEmptyArray verifies a nil order marshals as a
// JSON array (not null), so the wire body is always well-formed.
func TestReorderShortcutsNilEncodesEmptyArray(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ReorderShortcuts(nil); err != nil {
		t.Fatalf("ReorderShortcuts error: %v", err)
	}
	if !containsAll(string(rec.Body), `"order":[]`) {
		t.Errorf("body = %q", string(rec.Body))
	}
}
