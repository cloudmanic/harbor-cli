// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// GetSettings fetches the authenticated user's effective account preferences
// (theme, default notebook, sort, locale, timezone, and the nested
// notification_prefs / editor_prefs objects), data-wrapped. The stored document
// is overlaid on code-owned defaults, so the response is always complete even
// before the first write. Settings are NOT part of the sync stream.
func (c *Client) GetSettings() ([]byte, error) {
	return c.doGet("/settings", nil)
}

// UpdateSettings applies a partial (or full) preferences update. Provided
// top-level scalars overwrite; the two nested objects (notification_prefs,
// editor_prefs) deep-merge field by field on the server, so a client can set
// just one nested sub-field without clearing the rest. default_notebook_id is
// nullable: omit to leave unchanged, send a string to adopt, or an explicit nil
// to clear. Returns the full merged preferences object (data-wrapped).
func (c *Client) UpdateSettings(body map[string]any) ([]byte, error) {
	return c.doPut("/settings", body)
}
