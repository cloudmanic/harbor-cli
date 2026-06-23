// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"strings"
	"testing"
)

func TestDisplaySessions(t *testing.T) {
	data := []byte(`{"data":[
		{"id":"fam1","device_id":"ios-9F3A","device_name":"Jane's iPhone","ip":"203.0.113.7","last_seen_at":1750000000000,"current":true},
		{"id":"fam2","device_id":"web","device_name":"Chrome","ip":"203.0.113.8","last_seen_at":1750000000000,"current":false}
	],"paging":{"offset":0,"total":2}}`)
	out := captureStdout(t, func() { displaySessions(data) })
	if !strings.Contains(out, "fam1") || !strings.Contains(out, "Jane's iPhone") {
		t.Errorf("sessions table missing fields:\n%s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("current session mark missing:\n%s", out)
	}
}
