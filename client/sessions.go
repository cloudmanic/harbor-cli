// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

// ListSessions lists the user's active sessions (one per refresh-token family),
// marking the current one. Standard collection envelope.
func (c *Client) ListSessions(params map[string]string) ([]byte, error) {
	return c.doGet("/sessions", params)
}

// RevokeSession revokes a single session (family) by its id.
func (c *Client) RevokeSession(id string) ([]byte, error) {
	return c.doDelete("/sessions/"+id, nil)
}

// RevokeSessions revokes multiple sessions: every session (except == "") or all
// but the current one (except == "current").
func (c *Client) RevokeSessions(except string) ([]byte, error) {
	params := map[string]string{}
	if except != "" {
		params["except"] = except
	}
	return c.doDelete("/sessions", params)
}
