// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListShortcuts returns the user's sidebar shortcuts (collection envelope),
// ordered by position ascending. Accepts the standard list params (limit,
// offset, order, include_deleted).
func (c *Client) ListShortcuts(params map[string]string) ([]byte, error) {
	return c.doGet("/shortcuts", params)
}

// GetShortcut fetches a single shortcut by id. With includeDeleted, a
// tombstoned shortcut is returned instead of a 404.
func (c *Client) GetShortcut(id string, includeDeleted bool) ([]byte, error) {
	params := map[string]string{}
	if includeDeleted {
		params["include_deleted"] = "true"
	}
	return c.doGet("/shortcuts/"+id, params)
}

// CreateShortcut creates a shortcut from the given fields and returns the bare
// created object (201).
func (c *Client) CreateShortcut(body map[string]any) ([]byte, error) {
	return c.doPost("/shortcuts", body)
}

// UpdateShortcut applies a partial update (only the fields present in body) and
// returns the bare updated object.
func (c *Client) UpdateShortcut(id string, body map[string]any) ([]byte, error) {
	return c.doPatch("/shortcuts/"+id, body)
}

// DeleteShortcut tombstones a shortcut (204) so the removal propagates via sync.
func (c *Client) DeleteShortcut(id string) ([]byte, error) {
	return c.doDelete("/shortcuts/"+id, nil)
}

// ReorderShortcuts bulk-reorders the user's shortcuts. The order slice must
// reference every live shortcut id exactly once; the server renumbers to clean
// integer gaps and returns the reordered collection.
func (c *Client) ReorderShortcuts(order []string) ([]byte, error) {
	if order == nil {
		order = []string{}
	}
	return c.doPut("/shortcuts/order", map[string]any{"order": order})
}
