// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import "testing"

func TestSearch(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).Search(map[string]string{"q": "tag:finance budget", "types": "note"}); err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if rec.Path != "/search" {
		t.Errorf("path = %s", rec.Path)
	}
	if !containsAll(rec.Query, "q=tag", "types=note") {
		t.Errorf("query = %q", rec.Query)
	}
}

func TestSearchCoordinates(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"resource_id":"r1","pages":[]}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).SearchCoordinates(map[string]string{"resource_id": "r1", "terms": "budget,q3"}); err != nil {
		t.Fatalf("SearchCoordinates error: %v", err)
	}
	if rec.Path != "/search/coordinates" {
		t.Errorf("path = %s", rec.Path)
	}
	if !containsAll(rec.Query, "resource_id=r1", "terms=budget") {
		t.Errorf("query = %q", rec.Query)
	}
}
