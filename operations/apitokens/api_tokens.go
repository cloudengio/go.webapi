// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
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

// T represents a token that can be used to authenticate with an API.
// Tokens are typically of the form scheme://value where scheme is used
// to identify the reader that should be used to read the value.
type T struct {
	Scheme string
	Path   string
	value  []byte
}

// String returns a string representation of the token with the value
// redacted.
func (t T) String() string {
	return fmt.Sprintf("%v://****", t.Scheme)
}

// Token returns the value of the token.
func (t *T) Token() []byte {
	return t.value
}

// Read read's the token value using the supplied registry of Readers.
func (t *T) Read(ctx context.Context, registry *Readers) error {
	reader, ok := registry.Lookup(t.Scheme)
	if !ok {
		return fmt.Errorf("no reader for scheme: %v, expected one of: %v", t.Scheme, registry.Schemes())
	}
	val, err := reader.ReadFileCtx(ctx, t.Path)
	if err != nil {
		return err
	}
	t.value = val
	return nil
}

// Clone creates a copy of a Token that does not share any state with the original.
func (t *T) Clone() T {
	return T{
		Scheme: t.Scheme,
		Path:   t.Path,
		value:  slices.Clone(t.value),
	}
}

// New creates a token from the supplied text. If the text does not
// contain a scheme, the entire text is used as the value of the token
// and the scheme is the empty string. It is left to the caller to
// determine the appropriate interpretation of the value.
func New(text string) *T {
	idx := strings.Index(text, "://")
	if idx < 0 {
		return &T{Path: text, value: []byte(text)}
	}
	return &T{Scheme: text[:idx], Path: text[idx+3:]}
}

type tokenCtxKey int

var tokenCtxKeyVal tokenCtxKey

type tokenCtxStore struct {
	sync.Mutex
	tokens map[string]T
}

// ContextWithToken returns a new context that contains the provided
// named token in addition to any existing tokens.
func ContextWithToken(ctx context.Context, name string, token *T) context.Context {
	store := &tokenCtxStore{}
	if ostore, ok := ctx.Value(tokenCtxKeyVal).(*tokenCtxStore); ok {
		ostore.Lock()
		defer ostore.Unlock()
		store.tokens = maps.Clone(ostore.tokens)
	}
	if store.tokens == nil {
		store.tokens = make(map[string]T)
	}
	store.tokens[name] = token.Clone()
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
