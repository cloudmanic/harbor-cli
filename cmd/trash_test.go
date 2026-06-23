// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

// TestDisplayTrash checks the trash list renders titles and the paging footer.
func TestDisplayTrash(t *testing.T) {
	data := []byte(`{"data":[
		{"id":"n1","title":"Quarterly plan","notebook_id":"nb1aaaaaaa","is_encrypted":false,"trashed_at":1750000000000,"usn":90,"updated_at":1750000000000},
		{"id":"n2","title":"Old draft","notebook_id":"nb2","is_encrypted":true,"trashed_at":1750000000000,"usn":91,"updated_at":1750000000000}
	],"paging":{"limit":100,"offset":0,"total":2,"has_more":false}}`)
	out := captureStdout(t, func() { displayTrash(data) })
	if !strings.Contains(out, "Quarterly plan") || !strings.Contains(out, "Old draft") {
		t.Errorf("missing trashed note titles:\n%s", out)
	}
	if !strings.Contains(out, "showing 1–2 of 2") {
		t.Errorf("paging footer missing:\n%s", out)
	}
}

// TestDisplayTrashEmpty checks an empty trash renders the friendly no-results
// line rather than a bare table.
func TestDisplayTrashEmpty(t *testing.T) {
	data := []byte(`{"data":[],"paging":{"limit":100,"offset":0,"total":0,"has_more":false}}`)
	out := captureStdout(t, func() { displayTrash(data) })
	if !strings.Contains(out, "No results.") {
		t.Errorf("expected No results., got:\n%s", out)
	}
}

// TestDisplayRestoredNote checks a restore confirms and shows the note detail.
func TestDisplayRestoredNote(t *testing.T) {
	data := []byte(`{"id":"n1","title":"Quarterly plan","notebook_id":"nb1","in_trash":false,"is_encrypted":false,"usn":92,"updated_at":1750000060000}`)
	out := captureStdout(t, func() { displayRestoredNote(data) })
	if !strings.Contains(out, "restored") {
		t.Errorf("missing restored confirmation:\n%s", out)
	}
	if !strings.Contains(out, "Quarterly plan") || !strings.Contains(out, "n1") {
		t.Errorf("missing restored note fields:\n%s", out)
	}
}

// TestDisplayEmptyTrash checks the expunged-count message for plural and
// singular counts.
func TestDisplayEmptyTrash(t *testing.T) {
	out := captureStdout(t, func() { displayEmptyTrash([]byte(`{"expunged":3}`)) })
	if !strings.Contains(out, "3 notes") {
		t.Errorf("expected plural count, got:\n%s", out)
	}
	one := captureStdout(t, func() { displayEmptyTrash([]byte(`{"expunged":1}`)) })
	if !strings.Contains(one, "1 note permanently deleted") {
		t.Errorf("expected singular count, got:\n%s", one)
	}
}

// TestMapTrashError checks the trash-specific friendly error messages.
func TestMapTrashError(t *testing.T) {
	cases := map[string]string{
		"not_in_trash":      "not in the trash",
		"validation_failed": "invalid sort field",
	}
	for code, sub := range cases {
		got := mapTrashError(apiErr(code))
		if !strings.Contains(got.Error(), sub) {
			t.Errorf("mapTrashError(%s) = %q, want substring %q", code, got.Error(), sub)
		}
	}
}

// TestTrashConfirmEmptyYes verifies --yes bypasses the prompt entirely.
func TestTrashConfirmEmptyYes(t *testing.T) {
	if err := trashConfirmEmpty(true); err != nil {
		t.Errorf("with --yes, want nil, got %v", err)
	}
}

// TestTrashConfirmEmptyJSONRequiresYes verifies that in --json mode (where we
// must never prompt) confirmation without --yes is refused.
func TestTrashConfirmEmptyJSONRequiresYes(t *testing.T) {
	jsonOutput = true
	defer func() { jsonOutput = false }()
	err := trashConfirmEmpty(false)
	if err == nil {
		t.Fatal("expected refusal in --json mode without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("error = %q, want it to mention --yes", err.Error())
	}
}
