// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

func TestDisplayTags(t *testing.T) {
	data := []byte(`{"data":[
		{"id":"t1","name":"Receipts","parent_id":"","usn":12,"updated_at":1750000000000},
		{"id":"t2","name":"2026","parent_id":"t1xxxxxx","usn":13,"updated_at":1750000000000}
	],"paging":{"offset":0,"total":2}}`)
	out := captureStdout(t, func() { displayTags(data) })
	if !strings.Contains(out, "Receipts") || !strings.Contains(out, "2026") {
		t.Errorf("tag names missing:\n%s", out)
	}
	if !strings.Contains(out, "(top-level)") {
		t.Errorf("top-level marker missing:\n%s", out)
	}
}

func TestDisplayTagDetail(t *testing.T) {
	data := []byte(`{"id":"t1","name":"Receipts","parent_id":"","usn":12,"updated_at":1750000000000,"created_at":1749000000000}`)
	out := captureStdout(t, func() { displayTag(data) })
	if !strings.Contains(out, "Receipts") || !strings.Contains(out, "(top-level)") {
		t.Errorf("tag detail missing fields:\n%s", out)
	}
}

func TestMapTagError(t *testing.T) {
	if got := mapTagError(apiErr("tag_name_exists")); !strings.Contains(got.Error(), "already exists") {
		t.Errorf("tag_name_exists = %q", got.Error())
	}
	if got := mapTagError(apiErr("tag_cycle")); !strings.Contains(got.Error(), "cycle") {
		t.Errorf("tag_cycle = %q", got.Error())
	}
}
