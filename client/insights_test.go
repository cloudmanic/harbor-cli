// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"testing"
)

// TestListNoteLinks verifies the outgoing-links call uses GET on the right path
// and forwards the paging params.
func TestListNoteLinks(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ListNoteLinks("n1", map[string]string{"limit": "50", "offset": "10"}); err != nil {
		t.Fatalf("ListNoteLinks error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/notes/n1/links" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if !containsAll(rec.Query, "limit=50", "offset=10") {
		t.Errorf("query = %q", rec.Query)
	}
}

// TestListNoteBacklinks verifies the incoming-links call uses GET on the right
// path and forwards the paging params.
func TestListNoteBacklinks(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ListNoteBacklinks("n1", map[string]string{"limit": "25"}); err != nil {
		t.Fatalf("ListNoteBacklinks error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/notes/n1/backlinks" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if rec.Query != "limit=25" {
		t.Errorf("query = %q", rec.Query)
	}
}

// TestListNoteAudit verifies the audit call uses GET on the right path and
// forwards the order and action filter params.
func TestListNoteAudit(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	params := map[string]string{"order": "-created_at", "action": "delete"}
	if _, err := testClient(srv.URL).ListNoteAudit("n1", params); err != nil {
		t.Fatalf("ListNoteAudit error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/notes/n1/audit" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if !containsAll(rec.Query, "order=-created_at", "action=delete") {
		t.Errorf("query = %q", rec.Query)
	}
}
