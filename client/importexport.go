// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// ImportENEX uploads an Evernote .enex file via multipart and starts an import.
// It returns the data-wrapped job summary bytes plus the HTTP status: 201 when
// the import ran inline, 202 when it was enqueued as a background job. The
// caller distinguishes the two from the status (the body shape is identical).
// targetNotebookID and filename are optional form fields; empty values are
// dropped. The whole multipart body is buffered so it can be replayed after a
// transparent token refresh.
func (c *Client) ImportENEX(filename, targetNotebookID string, fileContent io.Reader) ([]byte, int, error) {
	full, err := c.buildURL("/import/enex", nil)
	if err != nil {
		return nil, 0, err
	}

	// Build the multipart body in memory: the optional text fields followed by
	// the .enex file part (the field name the server expects is "file").
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fields := map[string]string{}
	if filename != "" {
		fields["filename"] = filename
	}
	if targetNotebookID != "" {
		fields["target_notebook_id"] = targetNotebookID
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, 0, fmt.Errorf("failed to write form field %q: %w", k, err)
		}
	}
	// The multipart file part needs a filename; fall back to a sensible default
	// when the caller did not supply one (the server defaults to evernote.enex).
	partName := filename
	if partName == "" {
		partName = "evernote.enex"
	}
	fw, err := w.CreateFormFile("file", partName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create file part: %w", err)
	}
	if _, err := io.Copy(fw, fileContent); err != nil {
		return nil, 0, fmt.Errorf("failed to copy file content: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, 0, fmt.Errorf("failed to finalize multipart body: %w", err)
	}

	// requestWithStatus surfaces the 201-inline vs 202-async status the caller
	// needs (doMultipart would discard it).
	return c.requestWithStatus(http.MethodPost, full, buf.Bytes(), w.FormDataContentType(), true)
}

// ImportStatus polls an import job by id and returns the data-wrapped status
// document (live counters plus the per-note error list). A 404 surfaces as an
// APIError with code not_found when the job is unknown or not the caller's.
func (c *Client) ImportStatus(jobID string) ([]byte, error) {
	return c.doGet("/import/enex/"+jobID, nil)
}

// ExportENEX exports a notebook or an explicit note selection to a raw ENEX
// document. The response body is the .enex file bytes (not a JSON envelope), so
// it returns the live *http.Response for streaming — the caller MUST close the
// body, and reads the X-Skipped-Encrypted header off the response before
// draining it. Targeting is XOR: pass a non-empty notebookID OR a non-empty
// noteIDs slice (the server rejects both or neither with validation_failed).
// includeResources inlines each linked attachment as a base64 <resource> block.
func (c *Client) ExportENEX(notebookID string, noteIDs []string, includeResources bool) (*http.Response, error) {
	body := map[string]any{}
	if notebookID != "" {
		body["notebook_id"] = notebookID
	}
	if len(noteIDs) > 0 {
		body["note_ids"] = noteIDs
	}
	if includeResources {
		body["include_resources"] = true
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}
	full, err := c.buildURL("/export/enex", nil)
	if err != nil {
		return nil, err
	}
	return c.rawPost(full, raw, "application/json")
}

// rawPost performs a POST whose successful response is streamed back as a live
// *http.Response (the caller closes the body) — the streaming sibling of doPost,
// used by ENEX export which returns a raw file plus a header to read. On a
// non-2xx response it drains and closes the body and returns a decoded APIError,
// retrying once after a transparent token refresh on a 401 invalid_token.
func (c *Client) rawPost(fullURL string, body []byte, contentType string) (*http.Response, error) {
	return c.rawPostWithRefresh(fullURL, body, contentType, true)
}

// rawPostWithRefresh is rawPost's implementation; allowRefresh is set false on
// the single retry so a failed refresh can never loop.
func (c *Client) rawPostWithRefresh(fullURL string, body []byte, contentType string, allowRefresh bool) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, fullURL, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.setCommonHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.LastRequestID = resp.Header.Get("X-Request-Id")
		return resp, nil
	}

	errBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	c.LastRequestID = resp.Header.Get("X-Request-Id")
	apiErr := parseAPIError(errBody, resp.StatusCode, c.LastRequestID)
	if allowRefresh && resp.StatusCode == http.StatusUnauthorized && apiErr.Code == "invalid_token" && c.OnUnauthorized != nil {
		if newTok, ok := c.refresh(); ok {
			c.AccessToken = newTok
			return c.rawPostWithRefresh(fullURL, body, contentType, false)
		}
	}
	return nil, apiErr
}
