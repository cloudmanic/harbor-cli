// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListTrash returns the notes currently in the trash (collection envelope),
// most-recently-trashed first. Accepts the standard list params (limit, offset,
// order). Sortable orders: trashed_at, updated_at, created_at, title (prefix -
// for descending; the default is -trashed_at).
func (c *Client) ListTrash(params map[string]string) ([]byte, error) {
	return c.doGet("/trash", params)
}

// RestoreNote returns a note from the trash to the live set, clearing in_trash
// and bumping its USN. It returns the restored note object (bare, not wrapped
// in a data envelope).
func (c *Client) RestoreNote(id string) ([]byte, error) {
	return c.doPost("/notes/"+id+"/restore", nil)
}

// ExpungeNote permanently deletes a note: it promotes the note to the canonical
// sync tombstone (deleted=1) so the deletion propagates to every device. It
// works whether or not the note is currently in the trash and returns no body
// (204 No Content).
func (c *Client) ExpungeNote(id string) ([]byte, error) {
	return c.doPost("/notes/"+id+"/expunge", nil)
}

// EmptyTrash expunges every note currently in the trash. It returns the bare
// {"expunged": N} object reporting how many notes were permanently deleted.
func (c *Client) EmptyTrash() ([]byte, error) {
	return c.doDelete("/trash", nil)
}
