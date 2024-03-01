// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

import (
	"context"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	"cloudeng.io/file/localfs"
)

// T represents a token that can be used to authenticate with an API.
// Tokens are typically of the form scheme://value where scheme is used
// to identify the reader that should be used to read the value.
type T struct {
	Scheme string
	value  string
}

// String returns a string representation of the token with the value
// redacted.
func (t T) String() string {
	return fmt.Sprintf("%v://****", t.Scheme)
}

// Token returns the value of the token.
func (t *T) Token() string {
	return t.value
}

// Parse parses a token from the supplied text. If the text does not
// contain a scheme, the entire text is used as the value of the token
// and the scheme is the empty string. It is left to the caller to
// determine the appropriate interpretation of the value.
func Parse(text string) T {
	idx := strings.Index(text, "://")
	if idx < 0 {
		return T{value: text}
	}
	return T{Scheme: text[:idx], value: text[idx+3:]}
}

// Reader is the interface that must be implemented by any token reader.
// ReadFileCtx is called with the path of the token and should return
// the value of the token, that is, the scheme (eg. file://) should
// be stripped from the path before calling ReadFileCtx.
type Reader interface {
	ReadFileCtx(ctx context.Context, path string) ([]byte, error)
}

// Readers is a registry of token readers.
type Readers struct {
	mu      sync.Mutex
	readers map[string]Reader
}

// Register registers a reader for the specified scheme.
func (r *Readers) Register(scheme string, rd Reader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.readers == nil {
		r.readers = make(map[string]Reader)
	}
	r.readers[scheme] = rd
}

// Schemes returns the list of registered schemes.
func (r *Readers) Schemes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.readers))
	for k := range r.readers {
		keys = append(keys, k)
	}
	return keys
}

// GetToken returns the token for the supplied text using the supplied
// registry of readers. The token is parsed from the text and the
// appropriate reader is used to read the value of the token.
func GetToken(ctx context.Context, registry *Readers, text string) (T, error) {
	token := Parse(text)
	if len(token.Scheme) == 0 {
		return T{}, fmt.Errorf("no scheme in token, expected one of: %v", registry.Schemes())
	}
	r, ok := registry.readers[token.Scheme]
	if !ok {
		return T{}, fmt.Errorf("no reader for scheme: %v, expected one of: %v", token.Scheme, registry.Schemes())
	}
	val, err := r.ReadFileCtx(ctx, token.value)
	if err != nil {
		return T{}, err
	}
	token.value = string(val)
	return token, nil
}

// DefaultReaders is the default registry of readers, it includes
// readers for local files, literal values and environment variables.
var DefaultReaders = &Readers{}

func init() {
	DefaultReaders.Register("file", localfs.New())
	DefaultReaders.Register("literal", &LiteralReader{})
	DefaultReaders.Register("env", &EnvVarReader{})
}

// CloneReaders returns a copy of the supplied readers.
func CloneReaders(readers *Readers) *Readers {
	if readers == nil {
		return &Readers{}
	}
	return &Readers{readers: maps.Clone(readers.readers)}
}

// EnvVarReader implements a Reader for environment variables. The text
// should be 'env://<env-var-name>'.
type EnvVarReader struct{}

func (r *EnvVarReader) ReadFileCtx(_ context.Context, name string) ([]byte, error) {
	v := os.Getenv(name)
	if len(v) == 0 {
		return nil, fmt.Errorf("%v %w", name, os.ErrNotExist)
	}
	return []byte(v), nil
}

// LiteralReader implements a Reader for a literal token. The text
// should be 'literal://<token>'. ReadFile will return the text string
// it is called with.
type LiteralReader struct{}

func (*LiteralReader) ReadFileCtx(_ context.Context, text string) ([]byte, error) {
	return []byte(text), nil
}
