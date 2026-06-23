// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

func TestDisplayNotebooks(t *testing.T) {
	data := []byte(`{"data":[
		{"id":"nb1","name":"Work","stack":"Projects","is_default":false,"default_encrypt":false,"is_public":false,"usn":42,"updated_at":1750000000000},
		{"id":"nb2","name":"Inbox","is_default":true,"usn":1,"updated_at":1750000000000}
	],"paging":{"limit":100,"offset":0,"total":2,"has_more":false}}`)
	out := captureStdout(t, func() { displayNotebooks(data) })
	if !strings.Contains(out, "Work") || !strings.Contains(out, "Inbox") {
		t.Errorf("missing notebook names:\n%s", out)
	}
	if !strings.Contains(out, "★") {
		t.Errorf("default notebook star missing:\n%s", out)
	}
	if !strings.Contains(out, "showing 1–2 of 2") {
		t.Errorf("paging footer missing:\n%s", out)
	}
}

func TestDisplayNotebookDetail(t *testing.T) {
	data := []byte(`{"id":"nb1","name":"Work","stack":"Projects","is_default":false,"usn":42,"updated_at":1750000000000,"created_at":1749000000000}`)
	out := captureStdout(t, func() { displayNotebook(data) })
	if !strings.Contains(out, "nb1") || !strings.Contains(out, "Work") {
		t.Errorf("detail view missing fields:\n%s", out)
	}
}

func TestMapNotebookError(t *testing.T) {
	cases := map[string]string{
		"notebook_name_exists":  "already exists",
		"cannot_delete_default": "cannot be deleted",
		"cannot_unset_default":  "always be a default",
	}
	for code, sub := range cases {
		got := mapNotebookError(apiErr(code))
		if !strings.Contains(got.Error(), sub) {
			t.Errorf("mapNotebookError(%s) = %q, want substring %q", code, got.Error(), sub)
		}
	}
}
