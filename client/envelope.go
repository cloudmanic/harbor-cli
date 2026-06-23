// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import "encoding/json"

// Paging is the collection pagination block. Harbor uses offset mode
// ({limit, offset, total, has_more}) by default and cursor mode
// ({limit, next_cursor, has_more}) on opt-in endpoints.
type Paging struct {
	Limit      int64  `json:"limit"`
	Offset     int64  `json:"offset"`
	Total      int64  `json:"total"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// TokenResponse is the bare OAuth2 token object returned by POST /oauth/token
// (not wrapped in a data envelope, per the OAuth2 spec). expires_in is in
// seconds (the one exception to Harbor's epoch-ms convention).
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

// collectionEnvelope is the wire shape for a paged collection.
type collectionEnvelope struct {
	Data   json.RawMessage `json:"data"`
	Paging *Paging         `json:"paging"`
}

// dataEnvelope is the wire shape for a wrapped single resource ({ "data": … }).
type dataEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// DecodeToken parses the bare OAuth token object.
func DecodeToken(data []byte) (*TokenResponse, error) {
	var t TokenResponse
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// DecodePaging extracts the paging block from a collection response. The bool
// reports whether a paging block was present.
func DecodePaging(data []byte) (Paging, bool) {
	var env collectionEnvelope
	if err := json.Unmarshal(data, &env); err != nil || env.Paging == nil {
		return Paging{}, false
	}
	return *env.Paging, true
}

// UnwrapData returns the bytes of the "data" member when present (a wrapped or
// collection envelope), otherwise the original bytes unchanged (a bare
// resource). This lets display code treat all shapes uniformly.
func UnwrapData(data []byte) []byte {
	var env dataEnvelope
	if err := json.Unmarshal(data, &env); err == nil && len(env.Data) > 0 {
		return env.Data
	}
	return data
}

// CollectionItems returns the raw JSON elements of a collection's data array.
// Returns nil when data is absent or not an array.
func CollectionItems(data []byte) []json.RawMessage {
	var env collectionEnvelope
	if err := json.Unmarshal(data, &env); err != nil || len(env.Data) == 0 {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(env.Data, &items); err != nil {
		return nil
	}
	return items
}
