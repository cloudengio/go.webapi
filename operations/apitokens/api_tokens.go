// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package apitokens provides types and functions for managing API tokens
// and is built on top of the cmdutil/keys package and its InmemoryKeyStore.
package apitokens

import (
	"context"
	"fmt"
	"os"

	"cloudeng.io/cmdutil/keys"
	"golang.org/x/oauth2"
)

// T represents a token that can be used to authenticate with an API.
type T struct {
	ID    string
	value string
}

// String returns a string representation of the token with the value
// redacted.
func (t T) String() string {
	return fmt.Sprintf("%v://****", t.ID)
}

// Token returns the value of the token.
func (t *T) Token() string {
	return t.value
}

// ContextWithToken returns a new context that contains the provided
// named token in addition to any existing tokens.
func ContextWithToken(ctx context.Context, name string, token *T) context.Context {
	ims, ok := keys.KeyStoreFromContext(ctx)
	if !ok {
		ims = keys.NewInmemoryKeyStore()
		ctx = keys.ContextWithKeyStore(ctx, ims)
	}
	ims.AddKey(keys.KeyInfo{ID: name, Token: string(token.value)})
	return ctx
}

// TokenFromContext returns the token for the specified name, if any,
// that are stored in the context.
func TokenFromContext(ctx context.Context, name string) (T, bool) {
	ki, ok := keys.KeyInfoFromContextForID(ctx, name)
	if !ok {
		return T{}, false
	}
	return T{ID: ki.ID, value: ki.Token}, true
}

// TokenFromContextExpand returns the token for the specified name, if any,
// that are stored in the context. The token value will be expanded using
// ExpandEnv.
func TokenFromContextExpand(ctx context.Context, name string, mapping func(string) string) (T, bool) {
	ki, ok := keys.KeyInfoFromContextForID(ctx, name)
	if !ok {
		return T{}, false
	}
	return T{ID: ki.ID, value: ExpandEnv(ki.Token, mapping)}, true
}

// ContextWithOauth returns a new context that contains the provided
// named oauth2.TokenSource in addition to any existing TokenSources.
func ContextWithOAuth(ctx context.Context, name string, source oauth2.TokenSource) context.Context {
	if source == nil {
		return ctx // No-op if source is nil
	}
	ims, ok := keys.KeyStoreFromContext(ctx)
	if !ok {
		ims = keys.NewInmemoryKeyStore()
		ctx = keys.ContextWithKeyStore(ctx, ims)
	}
	ims.AddKey(keys.KeyInfo{ID: name, Extra: source})
	return ctx
}

// OAuthFromContext returns the TokenSource for the specified name, if any,
// that are stored in the context.
func OAuthFromContext(ctx context.Context, name string) oauth2.TokenSource {
	ki, ok := keys.KeyInfoFromContextForID(ctx, name)
	if !ok {
		return nil
	}
	if source, ok := ki.Extra.(oauth2.TokenSource); ok {
		return source
	}
	return nil
}

// ExpandEnv expands environment variables in the input string s using
// the provided mapping function. If mapping is nil, os.Getenv is used.
func ExpandEnv(s string, mapping func(string) string) string {
	if mapping == nil {
		mapping = os.Getenv
	}
	return os.Expand(s, mapping)
}

// New creates a new token with the specified ID and value.
func New(id, value string) *T {
	return &T{ID: id, value: value}
}
