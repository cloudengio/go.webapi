// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

import (
	"context"
	"encoding/json"
	"maps"
	"sync"

	"gopkg.in/yaml.v2"
)

type tokenKey int

var tokenKeyVal tokenKey

type tokenStore struct {
	sync.Mutex
	tokens map[string][]byte
}

// ContextWithTokens returns a new context that contains the provided
// named tokens in addition to any existing tokens. The tokens are typically
// encoded as JSON or YAML.
//
// Deprecated: use ContextWithToken instead.
func ContextWithTokens(ctx context.Context, name string, tokens []byte) context.Context {
	store := &tokenStore{
		tokens: map[string][]byte{name: tokens},
	}
	if ostore, ok := ctx.Value(tokenKeyVal).(*tokenStore); ok {
		ostore.Lock()
		defer ostore.Unlock()
		for k, v := range ostore.tokens {
			store.tokens[k] = v
		}
	}
	store.tokens[name] = tokens
	return context.WithValue(ctx, tokenKeyVal, store)
}

// TokensFromContext returns the tokens for the specified name, if any,
// that are stored in the context.
//
// Deprecated: use TokenFromContext instead.
func TokensFromContext(ctx context.Context, name string) ([]byte, bool) {
	if store, ok := ctx.Value(tokenKeyVal).(*tokenStore); ok {
		store.Lock()
		defer store.Unlock()
		t, ok := store.tokens[name]
		return t, ok
	}
	return nil, false
}

// ParseTokensYAML parses the tokens stored in the context for the specified
// name as YAML. It will return false if there are no tokens stored, true
// otherwise and an error if the unmarshal fails.
//
// Deprecated: use TokenFromContext instead.
func ParseTokensYAML(ctx context.Context, name string, cfg any) (bool, error) {
	tokens, ok := TokensFromContext(ctx, name)
	if !ok {
		return false, nil
	}
	return true, yaml.Unmarshal(tokens, cfg)
}

// ParseTokensJSON parses the tokens stored in the context for the specified
// name as JSON. It will return false if there are no tokens stored, true
// otherwise and an error if the unmarshal fails.
//
// Deprecated: use TokenFromContext instead.
func ParseTokensJSON(ctx context.Context, name string, cfg any) (bool, error) {
	tokens, ok := TokensFromContext(ctx, name)
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(tokens, cfg)
}

type tokenCtxKey int

var tokenCtxKeyVal tokenCtxKey

type tokenCtxStore struct {
	sync.Mutex
	tokens map[string]T
}

// ContextWithToken returns a new context that contains the provided
// named token in addition to any existing tokens.
func ContextWithToken(ctx context.Context, name string, token T) context.Context {
	store := &tokenCtxStore{}
	if ostore, ok := ctx.Value(tokenCtxKeyVal).(*tokenCtxStore); ok {
		ostore.Lock()
		defer ostore.Unlock()
		store.tokens = maps.Clone(ostore.tokens)
	}
	if store.tokens == nil {
		store.tokens = make(map[string]T)
	}
	store.tokens[name] = token
	return context.WithValue(ctx, tokenCtxKeyVal, store)
}

// TokenFromContext returns the token for the specified name, if any,
// that are stored in the context.
func TokenFromContext(ctx context.Context, name string) (T, bool) {
	if store, ok := ctx.Value(tokenCtxKeyVal).(*tokenCtxStore); ok {
		store.Lock()
		defer store.Unlock()
		t, ok := store.tokens[name]
		return t, ok
	}
	return T{}, false
}
