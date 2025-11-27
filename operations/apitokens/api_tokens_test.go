// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens_test

import (
	"context"
	"os"
	"testing"

	"cloudeng.io/webapi/operations/apitokens"
	"golang.org/x/oauth2"
)

func TestTokenContext(t *testing.T) {
	ctx := context.Background()

	t1 := apitokens.New("id1", "v1")
	t2 := apitokens.New("id2", "v2")

	if got, want := t1.String(), "id1://****"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	ctx = apitokens.ContextWithToken(ctx, "n1", t1)

	tok, ok := apitokens.TokenFromContext(ctx, "n1")
	if !ok {
		t.Errorf("expected token n1 to be present")
	}
	if got, want := tok.Token(), "v1"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	// Note: TokenFromContext returns a T where ID is the key name used in ContextWithToken
	if got, want := tok.ID, "n1"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	tok, ok = apitokens.TokenFromContext(ctx, "n2")
	if ok {
		t.Errorf("expected token n2 to be absent")
	}

	ctx = apitokens.ContextWithToken(ctx, "n2", t2)
	tok, ok = apitokens.TokenFromContext(ctx, "n1")
	if !ok {
		t.Errorf("expected token n1 to be present")
	}
	if got, want := tok.Token(), "v1"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	tok, ok = apitokens.TokenFromContext(ctx, "n2")
	if !ok {
		t.Errorf("expected token n2 to be present")
	}
	if got, want := tok.Token(), "v2"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := tok.ID, "n2"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTokenFromContextExpand(t *testing.T) {
	ctx := context.Background()
	t1 := apitokens.New("id1", `v1-${VAR}`)
	ctx = apitokens.ContextWithToken(ctx, "n1", t1)

	mapping := func(s string) string {
		if s == "VAR" {
			return "expanded"
		}
		return ""
	}

	tok, ok := apitokens.TokenFromContextExpand(ctx, "n1", mapping)
	if !ok {
		t.Errorf("expected token n1 to be present")
	}
	if got, want := tok.Token(), "v1-expanded"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "value")
	defer os.Unsetenv("TEST_VAR")

	if got, want := apitokens.ExpandEnv(`test-${TEST_VAR}`, nil), "test-value"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	mapping := func(s string) string {
		if s == "MY_VAR" {
			return "my-value"
		}
		return ""
	}
	if got, want := apitokens.ExpandEnv(`test-${MY_VAR}`, mapping), "test-my-value"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

type mockTokenSource struct {
	value string
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: m.value}, nil
}

func TestOAuthTokenSource(t *testing.T) {
	ctx := context.Background()

	ts := apitokens.OAuthFromContext(ctx, "n1")
	if ts != nil {
		t.Fatal("expected nil token source")
	}
	octx := ctx
	ctx = apitokens.ContextWithOAuth(ctx, "n1", nil)
	if ctx != octx {
		t.Fatal("expected context to be unchanged when setting nil token source")
	}

	ts = apitokens.OAuthFromContext(ctx, "n1")
	// Expect nil token source when set to nil.
	if ts != nil {
		t.Fatal("expected nil token source")
	}
	ts1 := &mockTokenSource{"ts1"}
	ts2 := &mockTokenSource{"ts2"}

	ctx = apitokens.ContextWithOAuth(ctx, "n1", ts1)
	ctx = apitokens.ContextWithOAuth(ctx, "n2", ts2)
	if got, want := apitokens.OAuthFromContext(ctx, "n1"), ts1; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := apitokens.OAuthFromContext(ctx, "n2"), ts2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
