// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// SyncPull returns server changes since a cursor. The body carries scope_id,
// after_usn, limit, and optionally device_id (advances that device's cursor).
func (c *Client) SyncPull(body map[string]any) ([]byte, error) {
	return c.doPost("/sync/pull", body)
}

// SyncPush uploads a batch of local change envelopes and returns per-change
// results. The body carries scope_id, device_id, and changes[].
func (c *Client) SyncPush(body map[string]any) ([]byte, error) {
	return c.doPost("/sync/push", body)
}

// RegisterDevice upserts a device by its client UUID and returns the bare
// device object.
func (c *Client) RegisterDevice(body map[string]any) ([]byte, error) {
	return c.doPost("/sync/devices", body)
}

// ListDevices lists the user's devices plus scope_max_usn and the GC floor.
// Note: this is a bare object with a top-level data array, not the standard
// {data, paging} collection envelope.
func (c *Client) ListDevices() ([]byte, error) {
	return c.doGet("/sync/devices", nil)
}

// RemoveDevice deregisters a device so it stops pinning the tombstone GC floor.
func (c *Client) RemoveDevice(deviceID string) ([]byte, error) {
	return c.doDelete("/sync/devices/"+deviceID, nil)
}

// SyncAck advances a device's last_acked_usn after a fully-applied pull chain.
func (c *Client) SyncAck(deviceID string, ackedUSN int64) ([]byte, error) {
	return c.doPost("/sync/ack", map[string]any{"device_id": deviceID, "acked_usn": ackedUSN})
}
