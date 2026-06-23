// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("abc"), 0644); err != nil {
		t.Fatal(err)
	}
	h, n, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile error: %v", err)
	}
	// sha256("abc") is a known constant.
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if h != want {
		t.Errorf("hash = %s, want %s", h, want)
	}
	if n != 3 {
		t.Errorf("size = %d, want 3", n)
	}
}

func TestWriteOutputToFileAndStdout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")
	n, err := writeOutput(path, bytes.NewReader([]byte("hello")))
	if err != nil || n != 5 {
		t.Fatalf("writeOutput file: n=%d err=%v", n, err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "hello" {
		t.Errorf("file content = %q", b)
	}

	out := captureStdout(t, func() {
		_, _ = writeOutput("-", bytes.NewReader([]byte("piped")))
	})
	if out != "piped" {
		t.Errorf("stdout content = %q", out)
	}
}

func TestFilenameFromContentDisposition(t *testing.T) {
	cases := map[string]string{
		`attachment; filename="diagram.png"`: "diagram.png",
		`attachment; filename=report.pdf`:    "report.pdf",
		`inline`:                             "",
	}
	for in, want := range cases {
		if got := filenameFromContentDisposition(in); got != want {
			t.Errorf("filenameFromContentDisposition(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDisplayFiles(t *testing.T) {
	data := []byte(`{"data":[{"hash":"e3b0c44298fc","mime":"image/png","size":1048576,"ocr_status":"done","thumb_status":"done","filename":"diagram.png","notes":[{"note_id":"a1"},{"note_id":"c3"}]}],"paging":{"offset":0,"total":1}}`)
	out := captureStdout(t, func() { displayFiles(data) })
	if !strings.Contains(out, "diagram.png") || !strings.Contains(out, "1.0 MB") {
		t.Errorf("files table missing fields:\n%s", out)
	}
	if !strings.Contains(out, "2") { // notes count
		t.Errorf("notes count missing:\n%s", out)
	}
}

func TestMapFileError(t *testing.T) {
	cases := map[string]string{
		"file_too_large":   "maximum upload size",
		"unsupported_type": "not allowed",
		"blob_missing":     "not stored",
	}
	for code, sub := range cases {
		if got := mapFileError(apiErr(code)); !strings.Contains(got.Error(), sub) {
			t.Errorf("mapFileError(%s) = %q", code, got.Error())
		}
	}
}
