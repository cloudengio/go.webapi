// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package operations provides support for invoking various operations
// on web APIs.
package operations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"cloudeng.io/net/ratecontrol"
)

// Encoding represents the encoding scheme used for the response body.
type Encoding int

const (
	JSONEncoding Encoding = iota
)

// Endpoint represents an API endpoint that whose response body is unmarshaled,
// by default using json.Unmarshal, into the specified type.
type Endpoint[T any] struct {
	options
}

// NewEndpoint returns a new endpoint for the specified type.
func NewEndpoint[T any](opts ...Option) *Endpoint[T] {
	ep := &Endpoint[T]{}
	for _, fn := range opts {
		fn(&ep.options)
	}
	if ep.rateController == nil {
		ep.rateController = ratecontrol.New()
	}
	if ep.unmarshal == nil {
		ep.unmarshal = json.Unmarshal
		ep.encoding = JSONEncoding
	}
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

func (ep *Endpoint[T]) isBackoffCode(code int) bool {
	for _, bc := range ep.backoffStatusCodes {
		if code == bc {
			return true
		}
	}
	return false
}

func (ep *Endpoint[T]) getWithResp(ctx context.Context, req *http.Request) (T, *http.Response, []byte, error) {
	var result T
	if err := ep.rateController.Wait(ctx); err != nil {
		return result, nil, nil, err
	}
	backoff := ep.rateController.Backoff()
	authSet := false
	for {
		select {
		case <-ctx.Done():
			return result, nil, nil, ctx.Err()
		default:
		}
		retries := backoff.Retries()
		var m T
		if !authSet && ep.auth != nil {
			if err := ep.auth.WithAuthorization(ctx, req); err != nil {
				return m, nil, nil, handleError(err, "", 0, retries)
			}
			authSet = true
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				return result, nil, nil, handleError(err, "", 0, retries)
			}
			log.Printf("network back off getting type: %T, retries: %v: %v", result, retries, err)
			if done, err := backoff.Wait(ctx, nil); done {
				return result, nil, nil, handleError(err, "", 0, retries)
			}
			continue
		}
		if ep.isBackoffCode(resp.StatusCode) {
			log.Printf("back off getting type: %T, retries: %v: %v", result, retries, resp.Status)
			if done, err := backoff.Wait(ctx, resp); done {
				return result, nil, nil, handleError(err, resp.Status, resp.StatusCode, retries)
			}
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return ep.handleResponse(resp, retries)
		}
		return m, nil, nil, handleError(nil, resp.Status, resp.StatusCode, retries)
	}
}

func (ep *Endpoint[T]) handleResponse(resp *http.Response, steps int) (T, *http.Response, []byte, error) {
	var result T
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, resp, body, handleError(err, resp.Status, resp.StatusCode, steps)
	}
	resp.Body.Close()
	err = ep.unmarshal(body, &result)
	return result, resp, body, handleError(err, resp.Status, resp.StatusCode, steps)
}
