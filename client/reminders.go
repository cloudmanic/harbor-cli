// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListReminders returns the user's reminders — live, non-trashed notes whose
// reminder_time is set — as a collection envelope. Accepts limit, offset,
// order, status, and due_before params (empty values are dropped by doGet).
func (c *Client) ListReminders(params map[string]string) ([]byte, error) {
	return c.doGet("/reminders", params)
}

// SetReminder sets or updates a note's reminder_time, returning the {note, usn}
// mutation envelope. The server treats PUT and POST equivalently for this
// route; we use PUT. The reminderTimeMS value is UTC epoch milliseconds.
func (c *Client) SetReminder(noteID string, reminderTimeMS int64) ([]byte, error) {
	return c.doPut("/notes/"+noteID+"/reminder", map[string]any{"reminder_time": reminderTimeMS})
}

// CompleteReminder marks a note's reminder done (sets reminder_done_time,
// defaulting to the server's current time), returning {note, usn}. When body
// is non-nil it may carry a done_time (epoch-ms) override.
func (c *Client) CompleteReminder(noteID string, body map[string]any) ([]byte, error) {
	if body == nil {
		body = map[string]any{}
	}
	return c.doPost("/notes/"+noteID+"/reminder/done", body)
}

// ClearReminder removes a note's reminder entirely (nulls both reminder_time
// and reminder_done_time), returning {note, usn}. It is idempotent.
func (c *Client) ClearReminder(noteID string) ([]byte, error) {
	return c.doDelete("/notes/"+noteID+"/reminder", nil)
}
