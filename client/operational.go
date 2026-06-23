// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import "net/http"

// Health probes the root /health liveness endpoint and returns the raw body.
//
// Liveness touches no dependency, so it returns 200 as long as the process can
// serve HTTP. Because the probe lives at the server root (no /api/v1 prefix),
// it is built from c.Origin() and dispatched through the low-level request
// method rather than doGet (which would prepend the versioned base URL). It is
// public, so allowRefresh is false — there is no token to refresh.
func (c *Client) Health() ([]byte, error) {
	return c.request(http.MethodGet, c.Origin()+"/health", nil, "", false)
}

// Ready probes the root /ready readiness endpoint and returns the raw body.
//
// Readiness checks every backing dependency (system DB, blob store) and returns
// 200 only when all pass, otherwise 503 with the same JSON shape. A 503 surfaces
// here as an *APIError, so callers that want the body on a not-ready response
// should inspect that error. Like /health it lives at the root and is public.
func (c *Client) Ready() ([]byte, error) {
	return c.request(http.MethodGet, c.Origin()+"/ready", nil, "", false)
}

// Version probes the root /version endpoint and returns the raw body of build
// metadata (version, commit, build_time, go_version). It lives at the server
// root and is public.
func (c *Client) Version() ([]byte, error) {
	return c.request(http.MethodGet, c.Origin()+"/version", nil, "", false)
}

// OpenAPI fetches the generated OpenAPI 3.0 document. Unlike the health probes
// this endpoint IS under the versioned API surface (/api/v1/openapi.json), so it
// goes through doGet, which prepends the /api/v1 base URL. It is public.
func (c *Client) OpenAPI() ([]byte, error) {
	return c.doGet("/openapi.json", nil)
}
