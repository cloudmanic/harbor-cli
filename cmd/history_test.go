// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

// TestDisplayHistory checks the list table shows version ids, the device
// placeholder for empty source devices, and the paging footer.
func TestDisplayHistory(t *testing.T) {
	data := []byte(`{"data":[
		{"id":"v1","note_id":"n1","usn_at_snapshot":88,"content_hash":"f1d2c3b4a5e6","is_encrypted":false,"source_device":"ipad-2","created_at":1750000000000},
		{"id":"v2","note_id":"n1","usn_at_snapshot":87,"content_hash":"00112233","is_encrypted":true,"created_at":1749000000000}
	],"paging":{"limit":100,"offset":0,"total":2,"has_more":false}}`)
	out := captureStdout(t, func() { displayHistory(data) })
	if !strings.Contains(out, "v1") || !strings.Contains(out, "v2") {
		t.Errorf("missing version ids:\n%s", out)
	}
	if !strings.Contains(out, "ipad-2") {
		t.Errorf("source device missing:\n%s", out)
	}
	if !strings.Contains(out, "—") {
		t.Errorf("empty-device placeholder missing:\n%s", out)
	}
	if !strings.Contains(out, "showing 1–2 of 2") {
		t.Errorf("paging footer missing:\n%s", out)
	}
}

// TestDisplayHistoryVersionMarkdown verifies the default (markdown) render
// strips HTML from the body and shows the snapshot metadata.
func TestDisplayHistoryVersionMarkdown(t *testing.T) {
	data := []byte(`{"id":"v1","note_id":"n1","title":"Quarterly plan","usn_at_snapshot":88,"content":"<p>Hello <strong>world</strong></p>","content_hash":"f1d2","is_encrypted":false,"source_device":"ipad-2","created_at":1750000000000}`)
	out := captureStdout(t, func() { displayHistoryVersion(data, "markdown") })
	if !strings.Contains(out, "Quarterly plan") || !strings.Contains(out, "v1") {
		t.Errorf("detail metadata missing:\n%s", out)
	}
	if !strings.Contains(out, "Hello world") {
		t.Errorf("HTML not stripped to text:\n%s", out)
	}
	if strings.Contains(out, "<strong>") {
		t.Errorf("raw HTML leaked in markdown mode:\n%s", out)
	}
}

// TestDisplayHistoryVersionHTML verifies the html format leaves the body markup
// intact.
func TestDisplayHistoryVersionHTML(t *testing.T) {
	data := []byte(`{"id":"v1","note_id":"n1","title":"Plan","content":"<p>Hello <strong>world</strong></p>","is_encrypted":false}`)
	out := captureStdout(t, func() { displayHistoryVersion(data, "html") })
	if !strings.Contains(out, "<strong>world</strong>") {
		t.Errorf("raw HTML missing in html mode:\n%s", out)
	}
}

// TestDisplayHistoryVersionEncrypted verifies encrypted snapshots render the
// placeholder instead of ciphertext.
func TestDisplayHistoryVersionEncrypted(t *testing.T) {
	data := []byte(`{"id":"v1","note_id":"n1","title":"x","content":"ciphertext-blob","is_encrypted":true}`)
	out := captureStdout(t, func() { displayHistoryVersion(data, "markdown") })
	if !strings.Contains(out, "[encrypted]") {
		t.Errorf("encrypted placeholder missing:\n%s", out)
	}
	if strings.Contains(out, "ciphertext-blob") {
		t.Errorf("ciphertext leaked:\n%s", out)
	}
}

// TestDisplayHistoryRevert verifies the {note, usn} envelope renders the
// restored note id, its new USN, and a confirmation line.
func TestDisplayHistoryRevert(t *testing.T) {
	data := []byte(`{"note":{"id":"n1","title":"Quarterly plan","notebook_id":"nb1","is_encrypted":false,"usn":91,"updated_at":1750000050000},"usn":91}`)
	out := captureStdout(t, func() { displayHistoryRevert(data) })
	if !strings.Contains(out, "n1") || !strings.Contains(out, "Quarterly plan") {
		t.Errorf("reverted note details missing:\n%s", out)
	}
	if !strings.Contains(out, "91") {
		t.Errorf("new USN missing:\n%s", out)
	}
	if !strings.Contains(out, "Reverted") {
		t.Errorf("confirmation line missing:\n%s", out)
	}
}

// TestMapHistoryError verifies the trash conflict maps to a friendly message
// and unrelated errors pass through unchanged.
func TestMapHistoryError(t *testing.T) {
	got := mapHistoryError(apiErr("note_in_trash"))
	if !strings.Contains(got.Error(), "restore it from trash") {
		t.Errorf("note_in_trash mapping = %q", got.Error())
	}
	passthrough := apiErr("not_found")
	if mapHistoryError(passthrough) != passthrough {
		t.Errorf("unrelated error should pass through unchanged")
	}
}
