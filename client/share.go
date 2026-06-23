// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"net/http"
)

// PublishShare publishes a note as a public, read-only page and returns the
// data-wrapped share object plus a fresh flag: true when a new share was minted
// (201 Created), false when the note was already public and the existing live
// share was returned unchanged (200 OK, idempotent). The body is optional and
// may carry a custom slug; pass nil to publish with a generated slug. Requires
// the notes scope.
func (c *Client) PublishShare(noteID string, body map[string]any) ([]byte, bool, error) {
	full, err := c.buildURL("/notes/"+noteID+"/share", nil)
	if err != nil {
		return nil, false, err
	}
	var raw []byte
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			return nil, false, err
		}
	}
	data, status, err := c.requestWithStatus(http.MethodPost, full, raw, "application/json", true)
	if err != nil {
		return nil, false, err
	}
	return data, status == http.StatusCreated, nil
}

// UnpublishShare revokes a note's public link. It is idempotent: a note that is
// already private, was never shared, or does not even exist still returns
// 204 No Content. Requires the notes scope.
func (c *Client) UnpublishShare(noteID string) ([]byte, error) {
	return c.doDelete("/notes/"+noteID+"/share", nil)
}

// PublicNote resolves and renders a shared note by its public token. This is a
// PUBLIC endpoint: it requires no bearer token or scope, so the calling client
// should be built with no credentials (newAnonymousClient). The response is the
// data-wrapped public render (title, content_html, attachments, view_count).
// Every failure mode collapses server-side to a single generic 404 not_found
// (anti-enumeration).
func (c *Client) PublicNote(token string) ([]byte, error) {
	return c.doGet("/public/notes/"+token, nil)
}
