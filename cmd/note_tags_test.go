// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	got := splitCSV(" a, b ,,c ")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("splitCSV = %v", got)
	}
	if g := splitCSV(""); len(g) != 0 {
		t.Errorf("splitCSV(empty) = %v, want empty slice", g)
	}
	if g := splitCSV("   "); len(g) != 0 {
		t.Errorf("splitCSV(spaces) = %v, want empty slice", g)
	}
}

func TestDisplayNoteTagJunctions(t *testing.T) {
	data := []byte(`{"data":[{"id":"j1","note_id":"n1xxxxxxx","tag_id":"t1","usn":92}],"paging":{}}`)
	out := captureStdout(t, func() { displayNoteTagJunctions(data) })
	if !strings.Contains(out, "j1") || !strings.Contains(out, "t1") {
		t.Errorf("junction table missing fields:\n%s", out)
	}
}

func TestDisplayNoteTagJunctionsEmpty(t *testing.T) {
	out := captureStdout(t, func() { displayNoteTagJunctions([]byte(`{"data":[],"paging":{}}`)) })
	if !strings.Contains(out, "no tags") {
		t.Errorf("expected empty message:\n%s", out)
	}
}
