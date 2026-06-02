// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
)

// PutEndpoint represents an API endpoint that supports PUT requests with a
// request body of type RequestT and a response body of type ResponseT.
type PutEndpoint[RequestT, ResponseT any] struct {
	options
}

func NewPutEndpoint[RequestT, ResponseT any](opts ...Option) *PutEndpoint[RequestT, ResponseT] {
	ep := &PutEndpoint[RequestT, ResponseT]{}
	handleOptions(&ep.options, opts...)
	return ep
}

// Put invokes a PUT request on this endpoint with a request of
// type RequestT and a response of type ResponseT.
func (ep *PutEndpoint[RequestT, ResponseT]) Put(ctx context.Context, url string, data RequestT) (ResponseT, []byte, Encoding, error) {
	return ep.putPost(ctx, url, "PUT", data)
}

// Post invokes a POST request on this endpoint with a request of
// type RequestT and a response of type ResponseT.
func (ep *PutEndpoint[RequestT, ResponseT]) Post(ctx context.Context, url string, data RequestT) (ResponseT, []byte, Encoding, error) {
	return ep.putPost(ctx, url, "POST", data)
}

func (ep *PutEndpoint[RequestT, ResponseT]) putPost(ctx context.Context, url string, method string, data RequestT) (ResponseT, []byte, Encoding, error) {
	var result ResponseT
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return result, nil, ep.encoding, err
	}
	if err := setRequestBody(req, ep.marshal, ep.marshalEncoding, data); err != nil {
		return result, nil, ep.encoding, err
	}
	result, _, body, err := issueRequest[ResponseT](ctx, ep.options, req)
	if err != nil {
		return result, body, ep.encoding, err
	}
	return result, body, ep.encoding, nil
}

// IssuePutPostRequest invokes an arbitrary PUT or POST request on this endpoint
// using the supplied http.Request except that the Request body is overridden
// with encoding of the supplied data.
func (ep *PutEndpoint[RequestT, ResponseT]) IssuePutPostRequest(ctx context.Context, req *http.Request, data RequestT) (ResponseT, []byte, Encoding, *http.Response, error) {
	var result ResponseT
	if req.Method != "PUT" && req.Method != "POST" {
		return result, nil, ep.encoding, nil, errors.New("only PUT or POST methods are supported")
	}
	if err := setRequestBody(req, ep.marshal, ep.marshalEncoding, data); err != nil {
		return result, nil, ep.encoding, nil, err
	}
	t, r, b, err := issueRequest[ResponseT](ctx, ep.options, req)
	return t, b, ep.encoding, r, err
}

func setRequestBody[RequestT any](req *http.Request, m Marshal, e Encoding, data RequestT) error {
	reqBody, err := m(data)
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(reqBody))
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("Content-Type", e.ContentType())
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(reqBody)), nil
	}
	return nil
}
