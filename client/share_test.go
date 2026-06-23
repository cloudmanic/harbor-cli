// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

// TestPublishShareFreshVsExisting verifies the POST path/body and that the
// 201/200 status maps to fresh=true/false (idempotent publish).
func TestPublishShareFreshVsExisting(t *testing.T) {
	// 201 → fresh=true, with a custom slug in the body.
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"data":{"note_id":"n1","share_token":"Xa9","slug":"my-slug","public_url":"http://localhost:3000/p/Xa9","is_public":true}}`)
	data, fresh, err := testClient(srv.URL).PublishShare("n1", map[string]any{"slug": "my-slug"})
	srv.Close()
	if err != nil {
		t.Fatalf("PublishShare error: %v", err)
	}
	if !fresh {
		t.Error("expected fresh=true for 201")
	}
	if rec.Method != "POST" || rec.Path != "/notes/n1/share" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if rec.ContentType != "application/json" {
		t.Errorf("content-type = %q", rec.ContentType)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["slug"] != "my-slug" {
		t.Errorf("body slug = %v", body["slug"])
	}
	if !containsAll(string(data), "public_url", "my-slug") {
		t.Errorf("response = %s", data)
	}

	// 200 → fresh=false (already public, idempotent).
	srv2 := newTestServer(t, nil, 200, `{"data":{"note_id":"n1","is_public":true}}`)
	defer srv2.Close()
	_, fresh2, err := testClient(srv2.URL).PublishShare("n1", nil)
	if err != nil {
		t.Fatalf("PublishShare(existing) error: %v", err)
	}
	if fresh2 {
		t.Error("expected fresh=false for 200")
	}
}

// TestPublishShareNoBodyOmitsPayload verifies a nil body publishes with no
// request payload (a generated slug).
func TestPublishShareNoBodyOmitsPayload(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"data":{"note_id":"n1"}}`)
	defer srv.Close()
	if _, _, err := testClient(srv.URL).PublishShare("n1", nil); err != nil {
		t.Fatalf("PublishShare error: %v", err)
	}
	if len(rec.Body) != 0 {
		t.Errorf("expected empty body, got %q", rec.Body)
	}
}

// TestUnpublishShare verifies the DELETE method and path.
func TestUnpublishShare(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 204, ``)
	defer srv.Close()
	if _, err := testClient(srv.URL).UnpublishShare("n1"); err != nil {
		t.Fatalf("UnpublishShare error: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/notes/n1/share" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

// TestPublicNote verifies the GET path. It also asserts the request still works
// with no bearer token (an anonymous client), since this is a public endpoint.
func TestPublicNote(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"title":"Quarterly plan","content_html":"<p>Hi</p>","view_count":13}}`)
	defer srv.Close()
	// Build an anonymous client (no token) to mirror the public-endpoint usage.
	anon := NewClient(srv.URL, "")
	data, err := anon.PublicNote("Xa9Kd")
	if err != nil {
		t.Fatalf("PublicNote error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/public/notes/Xa9Kd" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if rec.Auth != "" {
		t.Errorf("expected no Authorization header on public read, got %q", rec.Auth)
	}
	if !containsAll(string(data), "Quarterly plan", "content_html") {
		t.Errorf("response = %s", data)
	}
}

// TestPublicNoteNotFound verifies the generic 404 (anti-enumeration) surfaces
// as an APIError, untranslated.
func TestPublicNoteNotFound(t *testing.T) {
	srv := newTestServer(t, nil, 404, `{"error":{"code":"not_found","message":"Not found."}}`)
	defer srv.Close()
	_, err := NewClient(srv.URL, "").PublicNote("nope")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("want *APIError, got %T", err)
	}
	if apiErr.Code != "not_found" || apiErr.Status != 404 {
		t.Errorf("apiErr = %+v", apiErr)
	}
}
