// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

func TestAttachTagCreatedVsExisting(t *testing.T) {
	// 201 → created=true
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"id":"j1","tag_id":"t1"}`)
	_, created, err := testClient(srv.URL).AttachTag("n1", map[string]any{"tag_name": "Receipts"})
	srv.Close()
	if err != nil {
		t.Fatalf("AttachTag error: %v", err)
	}
	if !created {
		t.Error("expected created=true for 201")
	}
	if rec.Path != "/notes/n1/tags" {
		t.Errorf("path = %s", rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["tag_name"] != "Receipts" {
		t.Errorf("body = %v", body)
	}

	// 200 → created=false (idempotent)
	srv2 := newTestServer(t, nil, 200, `{"id":"j1"}`)
	defer srv2.Close()
	_, created2, err := testClient(srv2.URL).AttachTag("n1", map[string]any{"tag_id": "t1"})
	if err != nil {
		t.Fatalf("AttachTag(existing) error: %v", err)
	}
	if created2 {
		t.Error("expected created=false for 200")
	}
}

func TestSetNoteTags(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).SetNoteTags("n1", []string{"t1", "t2"}); err != nil {
		t.Fatalf("SetNoteTags error: %v", err)
	}
	if rec.Method != "PUT" || rec.Path != "/notes/n1/tags" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	ids, _ := body["tag_ids"].([]any)
	if len(ids) != 2 {
		t.Errorf("tag_ids = %v", body["tag_ids"])
	}
}

func TestSetNoteTagsClearSendsEmptyArray(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).SetNoteTags("n1", nil); err != nil {
		t.Fatalf("SetNoteTags(nil) error: %v", err)
	}
	// nil must serialize as [] not null, so the server clears all tags.
	if !containsAll(string(rec.Body), `"tag_ids":[]`) {
		t.Errorf("body = %s, want tag_ids:[]", rec.Body)
	}
}

func TestDetachTag(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	defer srv.Close()
	if _, err := testClient(srv.URL).DetachTag("n1", "t1"); err != nil {
		t.Fatalf("DetachTag error: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/notes/n1/tags/t1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}
