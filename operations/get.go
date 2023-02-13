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
	"time"

	"cloudeng.io/net/ratecontrol"
)

// Endpoint represents an API endpoint that whose response body is unmarshaled,
// by default using json.Unmarshal, into the specified type.
type Endpoint[T any] struct {
	options
	ticker time.Ticker
	url    string
}

// NewEndpoint returns a new endpoint for the specified type and URL.
func NewEndpoint[T any](url string, opts ...Option) *Endpoint[T] {
	ep := &Endpoint[T]{url: url}
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
func (ep *Endpoint[T]) Get(ctx context.Context) (T, []byte, error) {
	return ep.get(ctx, ep.url, nil)
}

// Get invokes a Get request on this endpoint but with a body.
func (ep *Endpoint[T]) GetWithBody(ctx context.Context, body io.Reader) (T, []byte, error) {
	return ep.get(ctx, ep.url, body)
}

func (ep *Endpoint[T]) get(ctx context.Context, url string, body io.Reader) (T, []byte, error) {
	t, _, b, err := ep.getWithResp(ctx, ep.url, body)
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

func (ep *Endpoint[T]) getWithResp(ctx context.Context, url string, body io.Reader) (T, *http.Response, []byte, error) {
	var result T
	if err := ep.rateController.Wait(ctx); err != nil {
		return result, nil, nil, err
	}
	backoff := ep.rateController.Backoff()
	for {
		retries := backoff.Retries()
		var m T
		r, err := http.NewRequestWithContext(ctx, "GET", url, body)
		if err != nil {
			return m, nil, nil, handleError(err, "", 0, retries)
		}
		if ep.auth != nil {
			if err := ep.auth.WithAuthorization(ctx, r); err != nil {
				return m, nil, nil, handleError(err, "", 0, retries)
			}
		}
		resp, err := http.DefaultClient.Do(r)
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
