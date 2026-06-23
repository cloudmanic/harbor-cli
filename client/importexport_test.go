// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestImportENEXInlineVsAsync verifies the multipart upload hits the right path
// and that the returned status distinguishes an inline (201) from an enqueued
// (202) import.
func TestImportENEXInlineVsAsync(t *testing.T) {
	// 201 → inline.
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"data":{"import_job_id":"j1","status":"completed","total_notes":3,"imported_notes":3}}`)
	data, status, err := testClient(srv.URL).ImportENEX("my export.enex", "nb1", strings.NewReader("<en-export></en-export>"))
	srv.Close()
	if err != nil {
		t.Fatalf("ImportENEX error: %v", err)
	}
	if status != 201 {
		t.Errorf("status = %d, want 201 (inline)", status)
	}
	if rec.Method != "POST" || rec.Path != "/import/enex" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if !strings.HasPrefix(rec.ContentType, "multipart/form-data") {
		t.Errorf("content-type = %q", rec.ContentType)
	}
	body := string(rec.Body)
	// The file part is named "file"; the optional text fields ride alongside.
	if !containsAll(body, `name="file"`, "<en-export>", `name="filename"`, "my export.enex", `name="target_notebook_id"`, "nb1") {
		t.Errorf("multipart body missing expected parts:\n%s", body)
	}
	if !strings.Contains(string(data), "import_job_id") {
		t.Errorf("body = %s", data)
	}

	// 202 → enqueued.
	srv2 := newTestServer(t, nil, 202, `{"data":{"import_job_id":"j2","status":"queued"}}`)
	defer srv2.Close()
	_, status2, err := testClient(srv2.URL).ImportENEX("", "", strings.NewReader("<en-export></en-export>"))
	if err != nil {
		t.Fatalf("ImportENEX(async) error: %v", err)
	}
	if status2 != 202 {
		t.Errorf("status = %d, want 202 (async)", status2)
	}
}

// TestImportENEXOmitsEmptyFields verifies optional form fields are dropped when
// empty, and the file part still gets a default filename.
func TestImportENEXOmitsEmptyFields(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 201, `{"data":{}}`)
	defer srv.Close()
	if _, _, err := testClient(srv.URL).ImportENEX("", "", strings.NewReader("<en-export></en-export>")); err != nil {
		t.Fatalf("ImportENEX error: %v", err)
	}
	body := string(rec.Body)
	if strings.Contains(body, `name="filename"`) || strings.Contains(body, `name="target_notebook_id"`) {
		t.Errorf("empty optional fields should be omitted:\n%s", body)
	}
	// The file part must still carry a (default) filename for the server.
	if !containsAll(body, `name="file"`, "evernote.enex") {
		t.Errorf("default file part name missing:\n%s", body)
	}
}

// TestImportENEXTooLargeError verifies a 422 enex_too_large surfaces as an
// APIError with the right code.
func TestImportENEXTooLargeError(t *testing.T) {
	srv := newTestServer(t, nil, 422, `{"error":{"code":"enex_too_large","message":"too big"}}`)
	defer srv.Close()
	_, _, err := testClient(srv.URL).ImportENEX("big.enex", "", strings.NewReader("x"))
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != "enex_too_large" {
		t.Fatalf("want enex_too_large APIError, got %v", err)
	}
}

// TestImportStatus verifies the poll hits GET /import/enex/:id.
func TestImportStatus(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"id":"j1","status":"partial","errors":[]}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).ImportStatus("j1"); err != nil {
		t.Fatalf("ImportStatus error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/import/enex/j1" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

// TestExportENEXByNotebook verifies the export POST body and that the raw body
// streams back for a notebook target.
func TestExportENEXByNotebook(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `<?xml version="1.0"?><en-export></en-export>`)
	defer srv.Close()
	resp, err := testClient(srv.URL).ExportENEX("nb1", nil, true)
	if err != nil {
		t.Fatalf("ExportENEX error: %v", err)
	}
	defer resp.Body.Close()
	if rec.Method != "POST" || rec.Path != "/export/enex" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["notebook_id"] != "nb1" {
		t.Errorf("notebook_id = %v", body["notebook_id"])
	}
	if body["include_resources"] != true {
		t.Errorf("include_resources = %v", body["include_resources"])
	}
	if _, ok := body["note_ids"]; ok {
		t.Errorf("note_ids should be omitted when empty: %v", body["note_ids"])
	}
	raw, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(raw), "<en-export>") {
		t.Errorf("raw body = %q", raw)
	}
}

// TestExportENEXByNotes verifies the note-selection body and reading the
// X-Skipped-Encrypted header off the live response.
func TestExportENEXByNotes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Skipped-Encrypted", "2")
		w.Header().Set("Content-Type", "application/enex+xml")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`<en-export></en-export>`))
	}))
	defer srv.Close()
	resp, err := testClient(srv.URL).ExportENEX("", []string{"n1", "n2"}, false)
	if err != nil {
		t.Fatalf("ExportENEX error: %v", err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-Skipped-Encrypted") != "2" {
		t.Errorf("X-Skipped-Encrypted = %q, want 2", resp.Header.Get("X-Skipped-Encrypted"))
	}
}

// TestExportENEXNotFound verifies a 404 surfaces as an APIError (and the body is
// closed for us — we never receive a live response on error).
func TestExportENEXNotFound(t *testing.T) {
	srv := newTestServer(t, nil, 404, `{"error":{"code":"not_found","message":"no such notebook"}}`)
	defer srv.Close()
	resp, err := testClient(srv.URL).ExportENEX("missing", nil, false)
	if err == nil {
		if resp != nil {
			resp.Body.Close()
		}
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != "not_found" || apiErr.Status != 404 {
		t.Fatalf("want not_found APIError, got %v", err)
	}
}
