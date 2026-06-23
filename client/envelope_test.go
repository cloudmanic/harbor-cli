// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package client

import "testing"

func TestDecodeToken(t *testing.T) {
	tok, err := DecodeToken([]byte(`{"access_token":"at_1","refresh_token":"rt_1","token_type":"Bearer","expires_in":3600,"scope":"notes sync"}`))
	if err != nil {
		t.Fatalf("DecodeToken error: %v", err)
	}
	if tok.AccessToken != "at_1" || tok.RefreshToken != "rt_1" {
		t.Errorf("tokens = %+v", tok)
	}
	if tok.ExpiresIn != 3600 || tok.Scope != "notes sync" {
		t.Errorf("token meta = %+v", tok)
	}
}

func TestDecodePaging(t *testing.T) {
	p, ok := DecodePaging([]byte(`{"data":[{"id":"1"}],"paging":{"limit":100,"offset":0,"total":3,"has_more":true}}`))
	if !ok {
		t.Fatal("expected paging present")
	}
	if p.Total != 3 || !p.HasMore || p.Limit != 100 {
		t.Errorf("paging = %+v", p)
	}

	if _, ok := DecodePaging([]byte(`{"id":"bare"}`)); ok {
		t.Error("bare resource should report no paging")
	}
}

func TestUnwrapData(t *testing.T) {
	// Wrapped single resource.
	got := UnwrapData([]byte(`{"data":{"id":"n1","title":"Hi"}}`))
	if string(got) != `{"id":"n1","title":"Hi"}` {
		t.Errorf("UnwrapData wrapped = %s", got)
	}
	// Bare resource passes through unchanged.
	bare := []byte(`{"id":"n2"}`)
	if string(UnwrapData(bare)) != string(bare) {
		t.Errorf("UnwrapData bare = %s", UnwrapData(bare))
	}
}

func TestCollectionItems(t *testing.T) {
	items := CollectionItems([]byte(`{"data":[{"id":"a"},{"id":"b"}],"paging":{}}`))
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	if string(items[0]) != `{"id":"a"}` {
		t.Errorf("item0 = %s", items[0])
	}
	// A wrapped single (object data) is not an array → nil.
	if CollectionItems([]byte(`{"data":{"id":"x"}}`)) != nil {
		t.Error("object data should not decode as collection items")
	}
}
