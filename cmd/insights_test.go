// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestDisplayNoteLinks checks the outgoing-links table: a resolved target shows
// its title, while a broken edge shows the broken mark and an em-dash title.
func TestDisplayNoteLinks(t *testing.T) {
	data := []byte(`{"data":[
		{"target_note_id":"5b1f2c9a","broken":false,"target":{"id":"5b1f2c9a","title":"Roadmap","in_trash":false}},
		{"target_note_id":"4d3c2b1a","broken":false,"target":{"id":"4d3c2b1a","title":"Trashed plan","in_trash":true}},
		{"target_note_id":"0000ffff","broken":true,"target":null}
	],"paging":{"limit":100,"offset":0,"total":3,"has_more":false}}`)
	out := captureStdout(t, func() { displayNoteLinks(data) })
	for _, want := range []string{"TARGET-ID", "BROKEN", "TARGET-TITLE", "IN-TRASH", "Roadmap", "5b1f2c9a", "0000ffff"} {
		if !strings.Contains(out, want) {
			t.Errorf("links table missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "showing 1–3 of 3") {
		t.Errorf("paging footer missing:\n%s", out)
	}
}

// TestDisplayNoteLinksEmpty checks the empty collection renders the friendly
// no-results line rather than an empty table.
func TestDisplayNoteLinksEmpty(t *testing.T) {
	data := []byte(`{"data":[],"paging":{"limit":100,"offset":0,"total":0,"has_more":false}}`)
	out := captureStdout(t, func() { displayNoteLinks(data) })
	if !strings.Contains(out, "No results.") {
		t.Errorf("expected no-results line:\n%s", out)
	}
}

// TestDisplayNoteBacklinks checks the incoming-links table shows the source id
// and title columns.
func TestDisplayNoteBacklinks(t *testing.T) {
	data := []byte(`{"data":[
		{"source_note_id":"7a1b2c3d","source":{"id":"7a1b2c3d","title":"Weekly review","in_trash":false}}
	],"paging":{"limit":100,"offset":0,"total":1,"has_more":false}}`)
	out := captureStdout(t, func() { displayNoteBacklinks(data) })
	for _, want := range []string{"SOURCE-ID", "SOURCE-TITLE", "7a1b2c3d", "Weekly review"} {
		if !strings.Contains(out, want) {
			t.Errorf("backlinks table missing %q:\n%s", want, out)
		}
	}
}

// TestDisplayNoteAudit checks the audit table renders the action, device, USN,
// a compact metadata summary, and handles null metadata.
func TestDisplayNoteAudit(t *testing.T) {
	data := []byte(`{"data":[
		{"id":"a1","note_id":"n1","action":"delete","device_id":"ipad-2","usn":90,"metadata":{"kind":"trash"},"created_at":1750000000000},
		{"id":"a2","note_id":"n1","action":"create","device_id":"ipad-2","usn":41,"metadata":null,"created_at":1749000000000}
	],"paging":{"limit":100,"offset":0,"total":2,"has_more":false}}`)
	out := captureStdout(t, func() { displayNoteAudit(data) })
	for _, want := range []string{"ACTION", "DEVICE", "USN", "METADATA", "CREATED", "delete", "create", "ipad-2", "kind=trash"} {
		if !strings.Contains(out, want) {
			t.Errorf("audit table missing %q:\n%s", want, out)
		}
	}
}

// TestAuditMetaSummary checks the metadata summarizer: null/empty → dot, scalars
// inline, arrays as counts, nested objects as field counts, sorted keys.
func TestAuditMetaSummary(t *testing.T) {
	if got := auditMetaSummary(nil); got != "·" {
		t.Errorf("nil metadata = %q, want dot", got)
	}
	if got := auditMetaSummary(map[string]any{}); got != "·" {
		t.Errorf("empty metadata = %q, want dot", got)
	}
	got := auditMetaSummary(map[string]any{
		"to":      "nb-2",
		"from":    "nb-1",
		"added":   []any{"t1", "t2"},
		"changed": map[string]any{"title": true},
		"count":   float64(3),
	})
	// Keys are sorted: added, changed, count, from, to.
	want := "added=[2] changed={1} count=3 from=nb-1 to=nb-2"
	if got != want {
		t.Errorf("auditMetaSummary = %q, want %q", got, want)
	}
}

// TestNotesAuditParams checks flag validation: valid action/order pass through,
// while unknown values produce a friendly pre-flight error.
func TestNotesAuditParams(t *testing.T) {
	// A fresh command mirroring the audit flag set.
	newCmd := func() *cobra.Command {
		c := &cobra.Command{Use: "audit"}
		c.Flags().Int("limit", 0, "")
		c.Flags().Int("offset", 0, "")
		c.Flags().String("order", "", "")
		c.Flags().String("action", "", "")
		return c
	}

	// Valid: action + order + paging all forwarded.
	c := newCmd()
	_ = c.Flags().Set("action", "delete")
	_ = c.Flags().Set("order", "-usn")
	_ = c.Flags().Set("limit", "200")
	params, err := notesAuditParams(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params["action"] != "delete" || params["order"] != "-usn" || params["limit"] != "200" {
		t.Errorf("params = %v", params)
	}

	// Invalid action.
	c = newCmd()
	_ = c.Flags().Set("action", "bogus")
	if _, err := notesAuditParams(c); err == nil || !strings.Contains(err.Error(), "invalid --action") {
		t.Errorf("expected invalid action error, got %v", err)
	}

	// Invalid order field.
	c = newCmd()
	_ = c.Flags().Set("order", "title")
	if _, err := notesAuditParams(c); err == nil || !strings.Contains(err.Error(), "invalid --order") {
		t.Errorf("expected invalid order error, got %v", err)
	}

	// Unset flags → empty params (server defaults apply).
	c = newCmd()
	params, err = notesAuditParams(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

// TestMapInsightError checks the friendly remapping of the codes these read-only
// endpoints surface.
func TestMapInsightError(t *testing.T) {
	cases := map[string]string{
		"not_found":         "note not found",
		"validation_failed": "invalid query",
	}
	for code, sub := range cases {
		got := mapInsightError(apiErr(code))
		if !strings.Contains(got.Error(), sub) {
			t.Errorf("mapInsightError(%s) = %q, want substring %q", code, got.Error(), sub)
		}
	}
}
