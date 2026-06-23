// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"net/http"
)

// ListNoteTags lists the live tags attached to a note (collection of tag
// objects).
func (c *Client) ListNoteTags(noteID string, params map[string]string) ([]byte, error) {
	return c.doGet("/notes/"+noteID+"/tags", params)
}

// AttachTag attaches a tag to a note by id (tag_id) or by name (tag_name,
// creating the tag if missing). It returns the junction bytes and a created
// flag: true when a new junction was made (201), false when a live junction
// already existed (200, idempotent).
func (c *Client) AttachTag(noteID string, body map[string]any) ([]byte, bool, error) {
	full, err := c.buildURL("/notes/"+noteID+"/tags", nil)
	if err != nil {
		return nil, false, err
	}
	raw, _ := json.Marshal(body)
	data, status, err := c.requestWithStatus(http.MethodPost, full, raw, "application/json", true)
	if err != nil {
		return nil, false, err
	}
	return data, status == http.StatusCreated, nil
}

// SetNoteTags replaces a note's complete tag set with the given tag ids
// ([] clears all). Returns the resulting live junctions as a collection.
func (c *Client) SetNoteTags(noteID string, tagIDs []string) ([]byte, error) {
	if tagIDs == nil {
		tagIDs = []string{}
	}
	return c.doPut("/notes/"+noteID+"/tags", map[string]any{"tag_ids": tagIDs})
}

// DetachTag detaches a tag from a note (idempotent; 204 even if not linked).
func (c *Client) DetachTag(noteID, tagID string) ([]byte, error) {
	return c.doDelete("/notes/"+noteID+"/tags/"+tagID, nil)
}
