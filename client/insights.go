// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListNoteLinks returns a note's outgoing links — the harbor:note edges its body
// contains — as a collection envelope. Each row carries the target id, a broken
// flag, and (when the target resolves) a trimmed target summary. Accepts the
// standard list params (limit, offset). Links are derived and read-only; there
// is no write endpoint.
func (c *Client) ListNoteLinks(noteID string, params map[string]string) ([]byte, error) {
	return c.doGet("/notes/"+noteID+"/links", params)
}

// ListNoteBacklinks returns the live notes that link TO a note — its incoming
// edges — as a collection envelope. Trashed and expunged source notes are
// excluded by the server. Each row carries the source id and a trimmed source
// summary. Accepts the standard list params (limit, offset).
func (c *Client) ListNoteBacklinks(noteID string, params map[string]string) ([]byte, error) {
	return c.doGet("/notes/"+noteID+"/backlinks", params)
}

// ListNoteAudit returns a note's append-only audit log as a collection envelope,
// newest-first by default. The trail outlives the note (it records the deletion
// itself), so a trashed/expunged note still returns its events. Accepts the
// standard list params (limit, offset, order) plus an optional action filter.
func (c *Client) ListNoteAudit(noteID string, params map[string]string) ([]byte, error) {
	return c.doGet("/notes/"+noteID+"/audit", params)
}
