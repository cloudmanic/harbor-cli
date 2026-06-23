// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"testing"
)

func TestSyncPull(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"scope_id":"u1","scope_max_usn":7,"has_more":false,"chunk":[]}`)
	defer srv.Close()
	_, err := testClient(srv.URL).SyncPull(map[string]any{"scope_id": "u1", "after_usn": 0})
	if err != nil {
		t.Fatalf("SyncPull error: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/sync/pull" {
		t.Errorf("%s %s", rec.Method, rec.Path)
	}
}

func TestSyncPush(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"results":[]}`)
	defer srv.Close()
	_, err := testClient(srv.URL).SyncPush(map[string]any{"scope_id": "u1", "device_id": "d1", "changes": []any{}})
	if err != nil {
		t.Fatalf("SyncPush error: %v", err)
	}
	if rec.Path != "/sync/push" {
		t.Errorf("path = %s", rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["device_id"] != "d1" {
		t.Errorf("body = %v", body)
	}
}

func TestDeviceLifecycle(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"device_id":"d1"}`)
	defer srv.Close()
	c := testClient(srv.URL)

	if _, err := c.RegisterDevice(map[string]any{"device_id": "d1", "platform": "cli"}); err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}
	if rec.Method != "POST" || rec.Path != "/sync/devices" {
		t.Errorf("register: %s %s", rec.Method, rec.Path)
	}

	if _, err := c.ListDevices(); err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if rec.Method != "GET" || rec.Path != "/sync/devices" {
		t.Errorf("list: %s %s", rec.Method, rec.Path)
	}

	if _, err := c.RemoveDevice("d1"); err != nil {
		t.Fatalf("RemoveDevice: %v", err)
	}
	if rec.Method != "DELETE" || rec.Path != "/sync/devices/d1" {
		t.Errorf("remove: %s %s", rec.Method, rec.Path)
	}
}

func TestSyncAck(t *testing.T) {
	var rec recordedRequest
	srv := newTestServer(t, &rec, 200, `{"device_id":"d1","last_acked_usn":95}`)
	defer srv.Close()
	if _, err := testClient(srv.URL).SyncAck("d1", 95); err != nil {
		t.Fatalf("SyncAck error: %v", err)
	}
	if rec.Path != "/sync/ack" {
		t.Errorf("path = %s", rec.Path)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body, &body)
	if body["acked_usn"].(float64) != 95 {
		t.Errorf("acked_usn = %v", body["acked_usn"])
	}
}
