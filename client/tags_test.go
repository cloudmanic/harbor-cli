// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"net/url"
	"testing"
)

func TestListTagsTopLevelSendsEmptyParent(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	// Top-level mode: parent_id must be present-but-empty.
	q := url.Values{}
	q.Set("parent_id", "")
	if _, err := testClient(srv.URL).ListTags(q); err != nil {
		t.Fatalf("ListTags error: %v", err)
	}
	if rec.Query != "parent_id=" {
		t.Errorf("query = %q, want parent_id= (explicit empty)", rec.Query)
	}
}

func TestCreateTag(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"id":"t1","name":"Receipts"}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).CreateTag(map[string]any{"name": "Receipts"}); err != nil {
		t.Fatalf("CreateTag error: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/tags" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

func TestUpdateTag(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"id":"t1"}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).UpdateTag("t1", map[string]any{"parent_id": ""}); err != nil {
		t.Fatalf("UpdateTag error: %v", err)
	}
	if rec.Method != "PATCH" || rec.Path != "/tags/t1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

func TestDeleteTagChildrenPolicy(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	defer srv.Close()
	if _, err := testClient(srv.URL).DeleteTag("t1", "orphan"); err != nil {
		t.Fatalf("DeleteTag error: %v", err)
	}
	if rec.Query != "children=orphan" {
		t.Errorf("query = %q", rec.Query)
	}
}

func TestListTagNotes(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ListTagNotes("t1", map[string]string{"notebook_id": "nb1"}); err != nil {
		t.Fatalf("ListTagNotes error: %v", err)
	}
	if rec.Path != "/tags/t1/notes" {
		t.Errorf("path = %s", rec.Path)
	}
}
