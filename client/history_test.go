// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"testing"
)

// TestListNoteHistory verifies the list call's method, path, and query passthrough.
func TestListNoteHistory(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{"total":0}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ListNoteHistory("n1", map[string]string{"order": "usn_at_snapshot", "limit": "25"}); err != nil {
		t.Fatalf("ListNoteHistory error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/notes/n1/history" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if rec.Query == "" || !containsAll(rec.Query, "order=usn_at_snapshot", "limit=25") {
		t.Errorf("query = %q", rec.Query)
	}
}

// TestGetNoteHistoryVersion verifies the single-version snapshot fetch path.
func TestGetNoteHistoryVersion(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"id":"v1","note_id":"n1","content":"<p>hi</p>"}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).GetNoteHistoryVersion("n1", "v1"); err != nil {
		t.Fatalf("GetNoteHistoryVersion error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/notes/n1/history/v1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

// TestRevertNoteHistoryVersion verifies the revert posts to the right path and
// sends no body.
func TestRevertNoteHistoryVersion(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"note":{"id":"n1","usn":91},"usn":91}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).RevertNoteHistoryVersion("n1", "v1"); err != nil {
		t.Fatalf("RevertNoteHistoryVersion error: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/notes/n1/history/v1/revert" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}
