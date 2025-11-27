// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens

/*
// Reader is the interface that must be implemented by any token reader.
type Reader interface {
	ReadFileCtx(ctx context.Context, path string) ([]byte, error)
}

// DefaultReaders is the default registry of readers, it includes
// readers for local files, literal values and environment variables.
var DefaultReaders = &registry.T[Reader]{}

func init() {
	DefaultReaders.Register("file", NewLocalfs)
	DefaultReaders.Register("literal", NewLiteralReader)
	DefaultReaders.Register("env", NewEnvReader)
}

// NewLocalfs creates a Reader that reads tokens from the local
// filesystem.
func NewLocalfs(ctx context.Context, args ...any) (Reader, error) {
	return localfs.New(), nil
}

// NewEnvReader creates a Reader that reads tokens from environment variables.
func NewEnvReader(ctx context.Context, args ...any) (Reader, error) {
	return &EnvVarReader{}, nil
}

// NewLiteralReader creates a Reader that returns the literal value
// passed to ReadFileCtx.
func NewLiteralReader(ctx context.Context, args ...any) (Reader, error) {
	return &LiteralReader{}, nil
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
*/
