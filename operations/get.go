// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package operations provides support for invoking various operations
// on web APIs.
package operations

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"cloudeng.io/net/ratecontrol"
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
	}
	return ep
}

// Get invokes a GET request on this endpoint (without a body).
func (ep *Endpoint[T]) Get(ctx context.Context, url string) (T, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		var result T
		return result, nil, err
	}
	return ep.get(ctx, req)
}

// GetUsingRequest invokes a Get request on this endpoint using the
// supplied http.Request.
func (ep *Endpoint[T]) GetUsingRequest(ctx context.Context, req *http.Request) (T, []byte, error) {
	return ep.get(ctx, req)
}

func (ep *Endpoint[T]) get(ctx context.Context, req *http.Request) (T, []byte, error) {
	t, _, b, err := ep.getWithResp(ctx, req)
	return t, b, err
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
	for {
		retries := backoff.Retries()
		var m T
		if ep.auth != nil {
			if err := ep.auth.WithAuthorization(ctx, req); err != nil {
				return m, nil, nil, handleError(err, "", 0, retries)
			}
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return result, nil, nil, handleError(err, "", 0, retries)
		}
		if ep.isBackoffCode(resp.StatusCode) {
			if done, err := backoff.Wait(ctx); done {
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
