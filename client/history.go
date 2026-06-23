// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListNoteHistory returns a note's version history as a metadata-only
// collection envelope (no content/attributes bodies). Accepts the standard
// list params: limit, offset, and order (-created_at default; created_at and
// usn_at_snapshot are sortable).
func (c *Client) ListNoteHistory(noteID string, params map[string]string) ([]byte, error) {
	return c.doGet("/notes/"+noteID+"/history", params)
}

// GetNoteHistoryVersion fetches one version's full snapshot — including title,
// content, and attributes_json — returned bare (not wrapped in data). The
// version must belong to noteID or the server returns 404.
func (c *Client) GetNoteHistoryVersion(noteID, versionID string) ([]byte, error) {
	return c.doGet("/notes/"+noteID+"/history/"+versionID, nil)
}

// RevertNoteHistoryVersion restores a past version as a new current version and
// returns the {note, usn} mutation envelope (the same shape as the note write
// endpoints). The note must be live and not in the trash; the request has no
// body.
func (c *Client) RevertNoteHistoryVersion(noteID, versionID string) ([]byte, error) {
	return c.doPost("/notes/"+noteID+"/history/"+versionID+"/revert", nil)
}
