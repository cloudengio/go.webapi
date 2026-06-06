// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package operations provides support for invoking various operations
// on web APIs.
package operations

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Encoding represents the encoding scheme used for the response body.
type Encoding int

const (
	JSONEncoding Encoding = iota
)

// ContentType returns the content type associated with this encoding.
func (e Encoding) ContentType() string {
	switch e {
	case JSONEncoding:
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// Endpoint represents an API endpoint that can be invoked using GET.
// The response body is unmarshaled into the specified type T.
// Use PutEndpoint for operations where both the request and response bodies
// can be typed.
type Endpoint[T any] struct {
	options
}

// NewEndpoint returns a new endpoint for the specified type.
func NewEndpoint[T any](opts ...Option) *Endpoint[T] {
	ep := &Endpoint[T]{}
	handleOptions(&ep.options, false, opts...)
	return ep
}

// Get invokes a GET request on this endpoint (without a body).
func (ep *Endpoint[T]) Get(ctx context.Context, url string) (T, []byte, Encoding, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		var result T
		return result, nil, ep.encoding, err
	}
	return ep.get(ctx, req)
}

// IssueRequest invokes an arbitrary request on this endpoint using the
// supplied http.Request. The Body in the http.Response has already been
// read and its contents returned as the second return value.
func (ep *Endpoint[T]) IssueRequest(ctx context.Context, req *http.Request) (T, []byte, Encoding, *http.Response, error) {
	t, r, b, err := ep.getWithResp(ctx, req)
	return t, b, ep.encoding, r, err
}

func (ep *Endpoint[T]) get(ctx context.Context, req *http.Request) (T, []byte, Encoding, error) {
	t, _, b, err := ep.getWithResp(ctx, req)
	return t, b, ep.encoding, err
}

func (ep *Endpoint[T]) getWithResp(ctx context.Context, req *http.Request) (T, *http.Response, []byte, error) {
	return issueRequest[T](ctx, ep.options, req)
}

func issueRequest[T any](ctx context.Context, opts options, req *http.Request) (T, *http.Response, []byte, error) {
	opts.logger.Info("starting request", "method", req.Method, "url", req.URL.Redacted())
	var result T
	if err := opts.rateController.Wait(ctx); err != nil {
		return result, nil, nil, err
	}
	backoff := opts.rateController.Backoff()
	start := time.Now()
	authSet := false
	for {
		select {
		case <-ctx.Done():
			return result, nil, nil, ctx.Err()
		default:
		}
		retries := backoff.Retries()
		var m T
		if !authSet && opts.auth != nil {
			if err := opts.auth.WithAuthorization(ctx, req); err != nil {
				return m, nil, nil, handleError(err, "", 0, opts.isSuccessCode, retries)
			}
			authSet = true
		}
		if req.GetBody != nil {
			body, gerr := req.GetBody()
			if gerr != nil {
				return result, nil, nil, handleError(gerr, "", 0, opts.isSuccessCode, retries)
			}
			req.Body = body
		}
		resp, err := opts.client.Do(req)
		if err != nil {
			if !opts.isErrorRetryableAndLog(ctx, req, err) {
				return result, nil, nil, handleError(err, "", 0, opts.isSuccessCode, retries)
			}
			if done, _ := backoff.Wait(ctx, nil); done {
				logBackoff(ctx, "network backoff giving up", req, retries, time.Since(start), true, err)
				return result, nil, nil, handleError(err, "", 0, opts.isSuccessCode, retries)
			}
			logBackoff(ctx, "network backoff", req, retries, time.Since(start), false, err)
			continue
		}
		if opts.isBackoffCode(resp.StatusCode) {
			resp.Body.Close()
			if done, _ := backoff.Wait(ctx, resp); done {
				logBackoff(ctx, "application backoff giving up", req, retries, time.Since(start), true, err)
				return result, nil, nil, handleError(err, resp.Status, resp.StatusCode, opts.isSuccessCode, retries)
			}
			logBackoff(ctx, "application backoff", req, retries, time.Since(start), false, err)
			continue
		}
		if opts.isSuccessCode(resp.StatusCode) {
			opts.logger.Info("request successful", "method", req.Method, "url", req.URL.Redacted())
			return handleResponse[T](resp, opts.unmarshal, retries, opts.isSuccessCode)
		}
		return handleErrorResponse[T](resp, retries, opts.isSuccessCode)
	}
}

func handleErrorResponse[T any](resp *http.Response, steps int, isOk func(int) bool) (T, *http.Response, []byte, error) {
	var result T
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	return result, resp, body, handleError(err, resp.Status, resp.StatusCode, isOk, steps)
}

func handleResponse[T any](resp *http.Response, unmarshal Unmarshal, steps int, isOk func(int) bool) (T, *http.Response, []byte, error) {
	var result T
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, resp, body, handleError(err, resp.Status, resp.StatusCode, isOk, steps)
	}
	if len(body) > 0 {
		err = unmarshal(body, &result)
	}
	return result, resp, body, handleError(err, resp.Status, resp.StatusCode, isOk, steps)
}
