// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

// TestListReminders verifies the GET path and that filter params are forwarded.
func TestListReminders(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"data":[],"paging":{}}`)
	defer srv.Close()
	_, err := testClient(srv.URL).ListReminders(map[string]string{"status": "done", "due_before": "1750100000000", "order": "-updated_at"})
	if err != nil {
		t.Fatalf("ListReminders error: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/reminders" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	if !containsAll(rec.Query, "status=done", "due_before=1750100000000", "order=-updated_at") {
		t.Errorf("query = %q", rec.Query)
	}
}

// TestSetReminder verifies a PUT to the reminder route carrying reminder_time.
func TestSetReminder(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"note":{"id":"n1"},"usn":96}`)
	defer srv.Close()
	_, err := testClient(srv.URL).SetReminder("n1", 1750100000000)
	if err != nil {
		t.Fatalf("SetReminder error: %v", err)
	}
	if rec.Method != "PUT" || rec.Path != "/notes/n1/reminder" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body, &body); err != nil {
		t.Fatalf("body not JSON: %v (%s)", err, rec.Body)
	}
	// JSON numbers decode to float64.
	if body["reminder_time"] != float64(1750100000000) {
		t.Errorf("reminder_time = %v", body["reminder_time"])
	}
}

// TestCompleteReminderWithDoneTime verifies the done path and optional body.
func TestCompleteReminderWithDoneTime(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"note":{"id":"n1"},"usn":97}`)
	defer srv.Close()
	_, err := testClient(srv.URL).CompleteReminder("n1", map[string]any{"done_time": int64(1750090000000)})
	if err != nil {
		t.Fatalf("CompleteReminder error: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/notes/n1/reminder/done" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body, &body); err != nil {
		t.Fatalf("body not JSON: %v (%s)", err, rec.Body)
	}
	if body["done_time"] != float64(1750090000000) {
		t.Errorf("done_time = %v", body["done_time"])
	}
}

// TestCompleteReminderNilBody verifies a nil body still posts a valid JSON
// object (the server defaults done_time to now).
func TestCompleteReminderNilBody(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"note":{"id":"n1"},"usn":97}`)
	defer srv.Close()
	_, err := testClient(srv.URL).CompleteReminder("n1", nil)
	if err != nil {
		t.Fatalf("CompleteReminder error: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body, &body); err != nil {
		t.Fatalf("nil body should marshal to {}, got %q (%v)", rec.Body, err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty body, got %v", body)
	}
}

// TestClearReminder verifies a DELETE to the reminder route with no body.
func TestClearReminder(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"note":{"id":"n1"},"usn":98}`)
	defer srv.Close()
	_, err := testClient(srv.URL).ClearReminder("n1")
	if err != nil {
		t.Fatalf("ClearReminder error: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/notes/n1/reminder" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}
