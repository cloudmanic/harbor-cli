// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// APIError is the typed form of the Harbor error envelope:
//
//	{ "error": { "code", "message", "details", "request_id" } }
//
// Commands branch on Code (e.g. invalid_token triggers refresh; sync_conflict;
// validation_failed shows Details). It implements error so errors.As recovers
// the typed value from a wrapped error chain.
type APIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details"`
	RequestID string         `json:"request_id"`
	Status    int            `json:"-"`
}

// errorEnvelope is the wire shape carrying an APIError.
type errorEnvelope struct {
	Error *APIError `json:"error"`
}

// Error renders a concise single-line description. Richer rendering (bulleted
// details, dim request id) lives in the cmd display layer.
func (e *APIError) Error() string {
	if e == nil {
		return "unknown API error"
	}
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, msg)
	}
	return msg
}

// DetailLines returns the validation details as sorted "field: message" lines,
// for friendly multi-line rendering. Non-string detail values are stringified.
func (e *APIError) DetailLines() []string {
	if e == nil || len(e.Details) == 0 {
		return nil
	}
	lines := make([]string, 0, len(e.Details))
	for k, v := range e.Details {
		lines = append(lines, fmt.Sprintf("%s: %v", k, v))
	}
	sort.Strings(lines)
	return lines
}

// parseAPIError decodes a non-2xx response body into an APIError. When the body
// is not a well-formed error envelope (e.g. a proxy 502, or an empty body), it
// synthesizes an APIError from the HTTP status so callers always get a typed,
// non-nil error.
func parseAPIError(body []byte, status int, requestID string) *APIError {
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error != nil && (env.Error.Code != "" || env.Error.Message != "") {
		env.Error.Status = status
		if env.Error.RequestID == "" {
			env.Error.RequestID = requestID
		}
		return env.Error
	}

	// Fallback: not a recognizable envelope. Surface what we can.
	msg := strings.TrimSpace(string(body))
	if msg == "" || len(msg) > 500 {
		msg = http.StatusText(status)
	}
	return &APIError{
		Code:      fallbackCode(status),
		Message:   msg,
		Status:    status,
		RequestID: requestID,
	}
}

// fallbackCode maps an HTTP status to a stable code when the body carries none,
// mirroring the canonical code table for the common statuses.
func fallbackCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusRequestEntityTooLarge:
		return "payload_too_large"
	case http.StatusUnsupportedMediaType:
		return "unsupported_media"
	case http.StatusUnprocessableEntity:
		return "validation_failed"
	case http.StatusTooManyRequests:
		return "rate_limited"
	case http.StatusNotImplemented:
		return "not_implemented"
	case http.StatusServiceUnavailable:
		return "timeout"
	default:
		if status >= 500 {
			return "internal_error"
		}
		return "error"
	}
}
