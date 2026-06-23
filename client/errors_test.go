// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseAPIErrorEnvelope(t *testing.T) {
	body := []byte(`{"error":{"code":"validation_failed","message":"Bad.","details":{"email":"is required"},"request_id":"req_abc"}}`)
	e := parseAPIError(body, 422, "req_header")
	if e.Code != "validation_failed" {
		t.Errorf("code = %s", e.Code)
	}
	if e.Status != 422 {
		t.Errorf("status = %d", e.Status)
	}
	if e.RequestID != "req_abc" {
		t.Errorf("request_id = %s (body should win over header)", e.RequestID)
	}
	lines := e.DetailLines()
	if len(lines) != 1 || lines[0] != "email: is required" {
		t.Errorf("detail lines = %v", lines)
	}
}

func TestParseAPIErrorUsesHeaderRequestIDWhenBodyMissing(t *testing.T) {
	body := []byte(`{"error":{"code":"not_found","message":"Nope."}}`)
	e := parseAPIError(body, 404, "req_header")
	if e.RequestID != "req_header" {
		t.Errorf("request_id = %s, want header fallback", e.RequestID)
	}
}

func TestParseAPIErrorFallback(t *testing.T) {
	// Non-envelope body (e.g. a proxy error) → synthesize from status.
	e := parseAPIError([]byte("Bad Gateway"), 502, "req_x")
	if e.Code != "internal_error" {
		t.Errorf("code = %s, want internal_error for 5xx", e.Code)
	}
	if e.Status != 502 {
		t.Errorf("status = %d", e.Status)
	}

	e2 := parseAPIError([]byte(""), 401, "")
	if e2.Code != "unauthorized" {
		t.Errorf("code = %s, want unauthorized", e2.Code)
	}
}

func TestAPIErrorImplementsError(t *testing.T) {
	var err error = &APIError{Code: "conflict", Message: "Clash.", Status: 409}
	if err.Error() != "conflict: Clash." {
		t.Errorf("Error() = %q", err.Error())
	}

	wrapped := fmt.Errorf("while doing thing: %w", err)
	var apiErr *APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatal("errors.As should recover the APIError from a wrapped chain")
	}
	if apiErr.Code != "conflict" {
		t.Errorf("recovered code = %s", apiErr.Code)
	}
}
