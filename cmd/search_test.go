// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

func TestDisplaySearchMixedHits(t *testing.T) {
	data := []byte(`{"data":[
		{"type":"note","note_id":"9c2e","title":"Quarterly plan","snippet":"the <em>budget</em> for Q3","score":8.42},
		{"type":"attachment","resource_id":"0f9c","note_id":"9c2e","filename":"invoice.pdf","snippet":"total <em>budget</em>","score":3.07,"has_coordinates":true}
	],"paging":{"offset":0,"total":2}}`)
	out := captureStdout(t, func() { displaySearch(data) })
	if !strings.Contains(out, "Quarterly plan") || !strings.Contains(out, "invoice.pdf") {
		t.Errorf("hits missing:\n%s", out)
	}
	// Snippet <em> tags should be stripped for the table.
	if strings.Contains(out, "<em>") {
		t.Errorf("snippet HTML not stripped:\n%s", out)
	}
	if !strings.Contains(out, "budget") {
		t.Errorf("snippet text missing:\n%s", out)
	}
}

func TestDisplayCoordinates(t *testing.T) {
	data := []byte(`{"data":{"resource_id":"0f9c","mime":"application/pdf","page_count":3,"terms":["budget"],"truncated":false,"pages":[
		{"page":0,"page_width":1700,"page_height":2200,"matches":[{"term":"budget","word":"Budget","word_index":42,"box":{"x":510,"y":880,"w":120,"h":28},"confidence":0.98}]}
	]}}`)
	out := captureStdout(t, func() { displayCoordinates(data) })
	if !strings.Contains(out, "Budget") || !strings.Contains(out, "0f9c") {
		t.Errorf("coordinates missing fields:\n%s", out)
	}
}

func TestMapSearchError(t *testing.T) {
	if got := mapSearchError(apiErr("encrypted_not_searchable")); !strings.Contains(got.Error(), "never searchable") {
		t.Errorf("encrypted_not_searchable = %q", got.Error())
	}
	if got := mapSearchError(apiErr("ocr_not_ready")); !strings.Contains(got.Error(), "OCR") {
		t.Errorf("ocr_not_ready = %q", got.Error())
	}
}
