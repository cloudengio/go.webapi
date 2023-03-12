// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

import (
	"context"
	"sync"
)

type tokenKey int

var tokenKeyVal tokenKey

type tokenStore struct {
	sync.Mutex
	tokens map[string]string
}

// ContextWithToken returns a new context that contains the provided
// named token in addition to any existing tokens.
func ContextWithToken(ctx context.Context, name, token string) context.Context {
	store := &tokenStore{
		tokens: map[string]string{name: token},
	}
	if ostore, ok := ctx.Value(tokenKeyVal).(*tokenStore); ok {
		ostore.Lock()
		defer ostore.Unlock()
		for k, v := range ostore.tokens {
			store.tokens[k] = v
		}
	}
	store.tokens[name] = token
	return context.WithValue(ctx, tokenKeyVal, store)
}

// TokenFromContext returns the token for the specified name, if any,
// that is stored in the context.
func TokenFromContext(ctx context.Context, name string) (string, bool) {
	if store, ok := ctx.Value(tokenKeyVal).(*tokenStore); ok {
		store.Lock()
		defer store.Unlock()
		t, ok := store.tokens[name]
		return t, ok
	}
	return "", false
}
