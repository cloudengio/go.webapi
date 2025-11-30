// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package apitokens provides types and functions for managing API tokens
// and is built on top of the cmdutil/keys package and its InmemoryKeyStore.
package apitokens

import (
	"context"
	"fmt"
	"io/fs"

	"cloudeng.io/cmdutil/keys"
	"golang.org/x/oauth2"
)

// ContextWithKey returns a new context that contains the provided
// named key.Info in addition to any existing keys. It wraps keys.ContextWithKey.
func ContextWithKey(ctx context.Context, ki keys.Info) context.Context {
	return keys.ContextWithKey(ctx, ki)
}

// KeyFromContext retrieves the key.Info for the specified id from the context.
// It wraps keys.KeyInfoFromContextForID.
func KeyFromContext(ctx context.Context, id string) (keys.Info, bool) {
	return keys.KeyInfoFromContextForID(ctx, id)
}

// TokenFromContext retrieves the token for the specified id from the
// context. It returns the token as a *keys.Token and a boolean indicating
// whether the token was found.
func TokenFromContext(ctx context.Context, id string) (*keys.Token, bool) {
	ki, ok := keys.KeyInfoFromContextForID(ctx, id)
	if !ok {
		return nil, false
	}
	return ki.Token(), true
}

// ContextWithOauth returns a new context that contains the provided
// named oauth2.TokenSource in addition to any existing TokenSources.
func ContextWithOAuth(ctx context.Context, id, user string, source oauth2.TokenSource) context.Context {
	ki := keys.NewInfo(id, user, nil, source)
	return keys.ContextWithKey(ctx, ki)
}

// OAuthFromContext returns the TokenSource for the specified name, if any,
// that are stored in the context.
func OAuthFromContext(ctx context.Context, id string) oauth2.TokenSource {
	ki, ok := keys.KeyInfoFromContextForID(ctx, id)
	if !ok {
		return nil
	}
	if extra := ki.Extra(); extra != nil {
		if source, ok := extra.(oauth2.TokenSource); ok {
			return source
		}
	}
	return nil
}

// Error represents an error related to API tokens.
type Error struct {
	KeyID   string
	Service string
	Err     error
}

// Error implements the error interface.
func (e Error) Error() string {
	return fmt.Sprintf("api token error for key %q service %q: %v", e.KeyID, e.Service, e.Err)
}

// Unwrap supports error wrapping by returning the underlying error.
func (e Error) Unwrap() error {
	return e.Err
}

// NewErrNotFound creates a new Error indicating that the specified token
// was not found. Errors.Is(err, fs.ErrNotExist) can be used to check
// for this condition.
func NewErrNotFound(keyID, service string) error {
	return Error{
		KeyID:   keyID,
		Service: service,
		Err:     fmt.Errorf("token not found: %w", fs.ErrNotExist),
	}
}

// ClearToken overwrites the contents of the provided token byte slice
// with zeros.
func ClearToken(token []byte) {
	for i := range token {
		token[i] = 0
	}
}
