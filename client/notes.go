// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListNotes returns the user's notes (collection envelope). Accepts limit,
// offset, order, notebook_id, tag, updated_since, deleted, and fields params.
func (c *Client) ListNotes(params map[string]string) ([]byte, error) {
	return c.doGet("/notes", params)
}

// GetNote fetches one note. params may include deleted=true and
// format=markdown.
func (c *Client) GetNote(id string, params map[string]string) ([]byte, error) {
	return c.doGet("/notes/"+id, params)
}

// CreateNote creates a note and returns the {note, usn} mutation envelope.
func (c *Client) CreateNote(body map[string]any) ([]byte, error) {
	return c.doPost("/notes", body)
}

// UpdateNote applies a partial update and returns {note, usn}.
func (c *Client) UpdateNote(id string, body map[string]any) ([]byte, error) {
	return c.doPatch("/notes/"+id, body)
}

// AppendNote appends a fragment to the end of a note's body and returns
// {note, usn}. Encrypted notes are rejected by the server.
func (c *Client) AppendNote(id string, body map[string]any) ([]byte, error) {
	return c.doPost("/notes/"+id+"/append", body)
}

// DeleteNote trashes a note (recoverable) by default, or expunges it directly
// when permanent is true.
func (c *Client) DeleteNote(id string, permanent bool) ([]byte, error) {
	params := map[string]string{}
	if permanent {
		params["permanent"] = "true"
	}
	return c.doDelete("/notes/"+id, params)
}
