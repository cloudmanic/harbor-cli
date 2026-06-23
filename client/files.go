// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// ListFiles returns the user's files, each with its linked notes (collection
// envelope). Accepts mime, note_id, ocr_status, is_encrypted, updated_since,
// and the standard list params.
func (c *Client) ListFiles(params map[string]string) ([]byte, error) {
	return c.doGet("/files", params)
}

// CheckFile reports whether a blob already exists for the given sha256 hash.
// A zero size is omitted.
func (c *Client) CheckFile(hash string, size int64) ([]byte, error) {
	body := map[string]any{"hash": hash}
	if size > 0 {
		body["size"] = size
	}
	return c.doPost("/files/check", body)
}

// UploadFile uploads a file via the direct multipart endpoint. The server
// computes the sha256 and returns the bare resource object. mime/filename are
// optional (filename defaults to the base name). The CLI intentionally does NOT
// use the internal presign-upload/commit endpoints.
func (c *Client) UploadFile(path, mime, filename string, isEncrypted bool) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	if filename == "" {
		filename = filepath.Base(path)
	}
	fields := map[string]string{"filename": filename}
	if mime != "" {
		fields["mime"] = mime
	}
	if isEncrypted {
		fields["is_encrypted"] = "true"
	}
	return c.doMultipart("/files/upload", fields, "file", filename, f)
}

// GetFileDownload returns the presigned download URL and basic metadata for a
// blob (no bytes).
func (c *Client) GetFileDownload(hash string) ([]byte, error) {
	return c.doGet("/files/"+hash, nil)
}

// RawDownload streams the blob bytes straight through the API (the caller must
// close the response body). Used when a presigned URL cannot be followed.
func (c *Client) RawDownload(hash string) (*http.Response, error) {
	return c.doGetRaw("/files/"+hash+"/raw", nil)
}

// FetchURL performs an unauthenticated GET against a presigned URL (its
// credentials live in the query string), returning the live response for
// streaming. The caller must close the body.
func (c *Client) FetchURL(rawURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid download URL: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: HTTP %s", strconv.Itoa(resp.StatusCode))
	}
	return resp, nil
}
