// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

// TestSharePublishDisplayPromptsURL verifies the public URL is rendered
// prominently (it is the key output) and the headline reflects fresh vs
// already-public.
func TestSharePublishDisplayPromptsURL(t *testing.T) {
	data := []byte(`{"data":{"note_id":"n1","share_token":"Xa9Kd","slug":"quarterly-plan","public_url":"http://localhost:3000/p/Xa9Kd","is_public":true,"view_count":12,"created_at":1750000000000}}`)

	fresh := captureStdout(t, func() { sharePublishDisplay(true)(data) })
	if !strings.Contains(fresh, "http://localhost:3000/p/Xa9Kd") {
		t.Errorf("public_url missing from fresh publish:\n%s", fresh)
	}
	if !strings.Contains(fresh, "Published.") {
		t.Errorf("fresh headline missing:\n%s", fresh)
	}
	if !strings.Contains(fresh, "quarterly-plan") {
		t.Errorf("slug missing:\n%s", fresh)
	}

	existing := captureStdout(t, func() { sharePublishDisplay(false)(data) })
	if !strings.Contains(existing, "Already public") {
		t.Errorf("idempotent headline missing:\n%s", existing)
	}
	if !strings.Contains(existing, "http://localhost:3000/p/Xa9Kd") {
		t.Errorf("public_url missing from existing publish:\n%s", existing)
	}
}

// TestDisplayPublicNote verifies the public render shows the title, the body
// (HTML stripped), and attachment URLs.
func TestDisplayPublicNote(t *testing.T) {
	data := []byte(`{"data":{
		"title":"Quarterly plan",
		"content_html":"<p>Hello <strong>world</strong></p>",
		"author":"Jane Doe",
		"source_url":"https://example.com/clip",
		"created_at":1749000000000,
		"updated_at":1750000000000,
		"attachments":[
			{"resource_id":"r1","filename":"diagram.png","mime":"image/png","size":1048576,"url":"https://s3.example/blobs/abc?sig=xyz"}
		],
		"view_count":13
	}}`)
	out := captureStdout(t, func() { displayPublicNote(data) })
	if !strings.Contains(out, "Quarterly plan") {
		t.Errorf("title missing:\n%s", out)
	}
	if !strings.Contains(out, "Hello world") {
		t.Errorf("body not stripped to text:\n%s", out)
	}
	if strings.Contains(out, "<strong>") {
		t.Errorf("HTML tags leaked into output:\n%s", out)
	}
	if !strings.Contains(out, "Jane Doe") {
		t.Errorf("author missing:\n%s", out)
	}
	if !strings.Contains(out, "diagram.png") || !strings.Contains(out, "https://s3.example/blobs/abc?sig=xyz") {
		t.Errorf("attachment row missing:\n%s", out)
	}
}

// TestDisplayPublicNoteNoAttachments verifies a note with no attachments omits
// the attachments section entirely.
func TestDisplayPublicNoteNoAttachments(t *testing.T) {
	data := []byte(`{"data":{"title":"Plain","content_html":"<p>Body</p>","view_count":1}}`)
	out := captureStdout(t, func() { displayPublicNote(data) })
	if !strings.Contains(out, "Body") {
		t.Errorf("body missing:\n%s", out)
	}
	if strings.Contains(out, "Attachments:") {
		t.Errorf("attachments header should be absent:\n%s", out)
	}
}

// TestMapShareError verifies the share-specific codes get friendly messages and
// the generic not_found passes through untranslated (anti-enumeration).
func TestMapShareError(t *testing.T) {
	cases := map[string]string{
		"encrypted_cannot_share": "encrypted",
		"slug_taken":             "already in use",
	}
	for code, sub := range cases {
		got := mapShareError(apiErr(code))
		if !strings.Contains(got.Error(), sub) {
			t.Errorf("mapShareError(%s) = %q, want substring %q", code, got.Error(), sub)
		}
	}

	// not_found is left as-is so the honest generic server message shows.
	nf := apiErr("not_found")
	if got := mapShareError(nf); got != nf {
		t.Errorf("mapShareError(not_found) should pass through unchanged, got %v", got)
	}
}
