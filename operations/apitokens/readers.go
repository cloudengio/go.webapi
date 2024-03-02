// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

import (
	"context"
	"fmt"
	"maps"
	"os"
	"sync"

	"cloudeng.io/file/localfs"
)

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

// Lookup returns the reader for the specified scheme, if any.
func (r *Readers) Lookup(scheme string) (Reader, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rd, ok := r.readers[scheme]
	return rd, ok
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
