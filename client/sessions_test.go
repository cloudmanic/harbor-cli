// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import "testing"

func TestListSessions(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ListSessions(nil); err != nil {
		t.Fatalf("ListSessions error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/sessions" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

func TestRevokeSession(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	defer srv.Close()
	if _, err := testClient(srv.URL).RevokeSession("fam1"); err != nil {
		t.Fatalf("RevokeSession error: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/sessions/fam1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

func TestRevokeSessionsBulk(t *testing.T) {
	// revoke-others → ?except=current
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	if _, err := testClient(srv.URL).RevokeSessions("current"); err != nil {
		t.Fatalf("RevokeSessions(current) error: %v", err)
	}
	srv.Close()
	if rec.Query != "except=current" {
		t.Errorf("query = %q, want except=current", rec.Query)
	}

	// revoke-all → no query
	var rec2 recordedRequest
	srv2 := newTestServer(t, &rec2, 204, ``)
	defer srv2.Close()
	if _, err := testClient(srv2.URL).RevokeSessions(""); err != nil {
		t.Fatalf("RevokeSessions(all) error: %v", err)
	}
	if rec2.Query != "" {
		t.Errorf("query = %q, want empty (revoke all)", rec2.Query)
	}
}
