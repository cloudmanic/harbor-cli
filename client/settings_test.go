// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

// TestGetSettings verifies the GET method and relative path.
func TestGetSettings(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"theme":"system"}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).GetSettings(); err != nil {
		t.Fatalf("GetSettings error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/settings" {
		t.Errorf("%s %s, want GET /settings", rec.Method, rec.Path)
	}
}

// TestUpdateSettingsUsesPUT verifies the PUT method, path, and that the body is
// sent through verbatim (including a nested object).
func TestUpdateSettingsUsesPUT(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"theme":"dark"}}`)
	defer srv.Close()
	body := map[string]any{
		"theme":        "dark",
		"editor_prefs": map[string]any{"font_size": 18},
	}
	if _, err := testClient(srv.URL).UpdateSettings(body); err != nil {
		t.Fatalf("UpdateSettings error: %v", err)
	}
	if rec.Method != "PUT" || rec.Path != "/settings" {
		t.Errorf("%s %s, want PUT /settings", rec.Method, rec.Path)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body, &got); err != nil {
		t.Fatalf("body not JSON: %v (%s)", err, rec.Body)
	}
	if got["theme"] != "dark" {
		t.Errorf("body theme = %v, want dark", got["theme"])
	}
	editor, ok := got["editor_prefs"].(map[string]any)
	if !ok {
		t.Fatalf("editor_prefs missing or not an object: %v", got["editor_prefs"])
	}
	if editor["font_size"] != float64(18) {
		t.Errorf("editor_prefs.font_size = %v, want 18", editor["font_size"])
	}
}

// TestUpdateSettingsClearsDefaultNotebook verifies an explicit null is preserved
// over the wire (clears the default notebook).
func TestUpdateSettingsClearsDefaultNotebook(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":{"default_notebook_id":null}}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).UpdateSettings(map[string]any{"default_notebook_id": nil}); err != nil {
		t.Fatalf("UpdateSettings error: %v", err)
	}
	// json.Marshal renders a nil value as an explicit JSON null.
	if got := string(rec.Body); got != `{"default_notebook_id":null}` {
		t.Errorf("body = %s, want explicit null", got)
	}
}
