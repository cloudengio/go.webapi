// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens_test

import (
	"context"
	"testing"

	"cloudeng.io/cmdutil/keys"
	"cloudeng.io/webapi/operations/apitokens"
	"golang.org/x/oauth2"
)

func TestKeyContext(t *testing.T) {
	ctx := context.Background()

	k1 := keys.NewInfo("k1", "u1", []byte("t1"), nil)
	ctx = apitokens.ContextWithKey(ctx, k1)

	got, ok := apitokens.KeyFromContext(ctx, "k1")
	if !ok {
		t.Fatal("expected key k1 to be present")
	}
	if got.ID != "k1" {
		t.Errorf("got %v, want k1", got.ID)
	}
	if got.User != "u1" {
		t.Errorf("got %v, want u1", got.User)
	}
	if string(got.Token().Value()) != "t1" {
		t.Errorf("got %v, want t1", string(got.Token().Value()))
	}

	_, ok = apitokens.KeyFromContext(ctx, "k2")
	if ok {
		t.Fatal("expected key k2 to be absent")
	}
}

type mockTokenSource struct {
	value string
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: m.value}, nil
}

func TestOAuthContext(t *testing.T) {
	ctx := context.Background()

	ts1 := &mockTokenSource{"ts1"}
	ctx = apitokens.ContextWithOAuth(ctx, "o1", "u1", ts1)

	got := apitokens.OAuthFromContext(ctx, "o1")
	if got == nil {
		t.Fatal("expected token source o1 to be present")
	}

	// Verify we can get the token from the source
	tok, err := got.Token()
	if err != nil {
		t.Fatalf("failed to get token: %v", err)
	}
	if tok.AccessToken != "ts1" {
		t.Errorf("got %v, want ts1", tok.AccessToken)
	}

	got = apitokens.OAuthFromContext(ctx, "o2")
	if got != nil {
		t.Fatal("expected token source o2 to be absent")
	}

	// Verify underlying key info
	ki, ok := apitokens.KeyFromContext(ctx, "o1")
	if !ok {
		t.Fatal("expected key info for o1")
	}
	if ki.User != "u1" {
		t.Errorf("got %v, want u1", ki.User)
	}
}
