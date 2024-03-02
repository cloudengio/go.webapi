// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens_test

import (
	"bytes"
	"context"
	"testing"

	"cloudeng.io/webapi/operations/apitokens"
)

func TestAPITokens(t *testing.T) {
	ctx := context.Background()
	has := func(n string, v []byte) {
		token, ok := apitokens.TokensFromContext(ctx, n)
		if got, want := ok, true; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token, v; !bytes.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	notHas := func(n string) {
		token, ok := apitokens.TokensFromContext(ctx, n)
		if got, want := ok, false; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token, ([]byte)(nil); !bytes.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	ctx = apitokens.ContextWithTokens(ctx, "n1", []byte("v1"))
	has("n1", []byte("v1"))
	notHas("n2")
	ctx = apitokens.ContextWithTokens(ctx, "n2", []byte("v2"))
	has("n1", []byte("v1"))
	has("n2", []byte("v2"))
	notHas("n3")
}

func TestTokenContext(t *testing.T) {
	ctx := context.Background()
	has := func(n string, v []byte) {
		t.Helper()
		token, ok := apitokens.TokenFromContext(ctx, n)
		if got, want := ok, true; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token.Scheme, "literal"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token.Token(), v; !bytes.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	notHas := func(n string) {
		t.Helper()
		token, ok := apitokens.TokenFromContext(ctx, n)
		if got, want := ok, false; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token.Scheme, ""; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got := token.Token(); got != nil {
			t.Errorf("expected nil token, not: %v", got)
		}
	}

	t1 := apitokens.New("literal://v1")
	t2 := apitokens.New("literal://v2")
	if err := t1.Read(ctx, apitokens.DefaultReaders); err != nil {
		t.Fatalf("failed to read token: %v", err)
	}
	if err := t2.Read(ctx, apitokens.DefaultReaders); err != nil {
		t.Fatalf("failed to read token: %v", err)
	}

	ctx = apitokens.ContextWithToken(ctx, "n1", t1)
	has("n1", []byte("v1"))
	notHas("n2")
	ctx = apitokens.ContextWithToken(ctx, "n2", t2)
	has("n1", []byte("v1"))
	has("n2", []byte("v2"))
	notHas("n3")

	apitokens.CorruptToken(t1)
	has("n1", []byte("v1"))
	has("n2", []byte("v2"))
}
