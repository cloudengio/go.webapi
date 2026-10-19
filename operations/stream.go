// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"net/http"
)

// Decodable specifies the constraint for types that can unmarshal themselves
// from a jsontext.Decoder, matching the json.UnmarshalerFrom interface in
// encoding/json/v2.
type Decodable interface {
	json.UnmarshalerFrom
}

// StreamEndpoint represents an API endpoint that produces a continuous stream
// of unmarshaled instances of type T.
type StreamEndpoint[T Decodable] struct {
	options
}

// NewStreamEndpoint returns a new StreamEndpoint for the specified Decodable type.
func NewStreamEndpoint[T Decodable](opts ...Option) *StreamEndpoint[T] {
	ep := &StreamEndpoint[T]{}
	handleOptions(&ep.options, false, opts...)
	return ep
}

// Get invokes a GET request on url and returns a typed Stream.
func (ep *StreamEndpoint[T]) Get(ctx context.Context, url string) (*Stream[T], error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return ep.IssueStream(ctx, req)
}

// IssueStream issues req and, on success, returns a typed Stream.
func (ep *StreamEndpoint[T]) IssueStream(ctx context.Context, req *http.Request) (*Stream[T], error) {
	_, r, _, err := doRequest(ctx, &ep.options, req, streamResponse[T])
	if err != nil {
		return nil, err
	}
	return &Stream[T]{
		resp: r,
		dec:  jsontext.NewDecoder(r.Body, ep.decOpts...),
	}, nil
}

// Stream wraps an active HTTP response stream and decodes items of type T.
type Stream[T Decodable] struct {
	resp *http.Response
	dec  *jsontext.Decoder
}

// Next reads and unmarshals the next item from the stream into val.
// It returns io.EOF when the stream terminates cleanly.
func (s *Stream[T]) Next(val T) error {
	return val.UnmarshalJSONFrom(s.dec)
}

// Close closes the underlying HTTP response body.
func (s *Stream[T]) Close() error {
	if s.resp != nil && s.resp.Body != nil {
		return s.resp.Body.Close()
	}
	return nil
}

// Response returns the underlying *http.Response.
func (s *Stream[T]) Response() *http.Response {
	return s.resp
}

// Decoder returns the underlying *jsontext.Decoder.
func (s *Stream[T]) Decoder() *jsontext.Decoder {
	return s.dec
}
