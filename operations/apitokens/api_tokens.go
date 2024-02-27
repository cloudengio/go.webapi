// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

import (
	"context"
	"encoding/json"
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
func ParseTokensJSON(ctx context.Context, name string, cfg any) (bool, error) {
	tokens, ok := TokensFromContext(ctx, name)
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(tokens, cfg)
}
