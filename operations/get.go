// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package operations provides support for invoking various operations
// on web APIs.
package operations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloudeng.io/algo/ratecontrol"
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

// NewStream returns a new endpoint for streaming responses without requiring
// a type parameter.
func NewStream(opts ...Option) *Endpoint[any] {
	return NewEndpoint[any](opts...)
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

// IssueStream is like IssueRequest except that it returns as soon as the
// response headers have been received, with the Body in the http.Response
// unread, for responses that are streamed. The caller must close the Body.
// A response that is not a success is read, up to a limit, and closed as for
// IssueRequest, and is returned with its error.
//
// The request is made with ctx, which also governs the Body: cancelling it ends
// the stream. Only establishing the stream is rate controlled and retried; the
// rate controller's backoff does not apply to a stream that has been
// established and then ends, so a caller that re-establishes it in a loop should
// delay between attempts. The http.Client, see WithHTTPClient, must not have a
// Timeout since that includes the time spent reading the Body. The type
// parameter of the Endpoint is not used.
func (ep *Endpoint[T]) IssueStream(ctx context.Context, req *http.Request) (*http.Response, error) {
	_, r, _, err := doRequest(ctx, &ep.options, req, streamResponse[T])
	return r, err
}

func (ep *Endpoint[T]) get(ctx context.Context, req *http.Request) (T, []byte, Encoding, error) {
	t, _, b, err := ep.getWithResp(ctx, req)
	return t, b, ep.encoding, err
}

func (ep *Endpoint[T]) getWithResp(ctx context.Context, req *http.Request) (T, *http.Response, []byte, error) {
	return issueRequest[T](ctx, ep.options, req)
}

// issueRequest issues req and reads and unmarshals its response.
func issueRequest[T any](ctx context.Context, opts options, req *http.Request) (T, *http.Response, []byte, error) {
	return doRequest(ctx, &opts, req, func(resp *http.Response, retries int) (T, *http.Response, []byte, error) {
		return handleResponse[T](resp, opts.unmarshal, retries, opts.isSuccessCode)
	})
}

// responseHandler handles a successful response; retries is the number of
// retries that were needed.
type responseHandler[T any] func(resp *http.Response, retries int) (T, *http.Response, []byte, error)

// streamResponse returns the response with its body unread.
func streamResponse[T any](resp *http.Response, _ int) (T, *http.Response, []byte, error) {
	var result T
	return result, resp, nil, nil
}

// doRequest issues req, handling rate control, authorization and retries, and
// calls handle with the response if it is a success.
func doRequest[T any](ctx context.Context, opts *options, req *http.Request, handle responseHandler[T]) (T, *http.Response, []byte, error) {
	opts.logger.Info("starting request", "method", req.Method, "url", req.URL.Redacted())
	var result T
	if err := opts.rateController.Wait(ctx); err != nil {
		return result, nil, nil, err
	}
	req = req.Clone(ctx)
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
		if !authSet && opts.auth != nil {
			if err := opts.auth.WithAuthorization(ctx, req); err != nil {
				return result, nil, nil, handleError(err, "", 0, opts.isSuccessCode, retries)
			}
			authSet = true
		}
		if err := prepareRequestBody(req, retries); err != nil {
			return result, nil, nil, handleError(err, "", 0, opts.isSuccessCode, retries)
		}
		resp, err := opts.attemptRequest(ctx, req, backoff, retries, start)
		if err != nil {
			return result, nil, nil, err
		}
		if resp == nil {
			continue
		}
		if opts.isSuccessCode(resp.StatusCode) {
			opts.logger.Info("request successful", "method", req.Method, "url", req.URL.Redacted())
			return handle(resp, retries)
		}
		return handleErrorResponse[T](resp, retries, opts.isSuccessCode)
	}
}

func prepareRequestBody(req *http.Request, retries int) error {
	if retries > 0 && req.Body != nil && req.GetBody == nil {
		return fmt.Errorf("request body cannot be rewound for retry: GetBody is nil")
	}
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return err
		}
		req.Body = body
	}
	return nil
}

func (opts *options) attemptRequest(ctx context.Context, req *http.Request, backoff ratecontrol.Backoff, retries int, start time.Time) (*http.Response, error) {
	resp, err := opts.client.Do(req)
	if err != nil {
		if !opts.isErrorRetryableAndLog(ctx, req, err) {
			return nil, handleError(err, "", 0, opts.isSuccessCode, retries)
		}
		if done, _ := backoff.Wait(ctx, nil); done {
			logBackoff(ctx, "network backoff giving up", req, retries, time.Since(start), true, err)
			return nil, handleError(err, "", 0, opts.isSuccessCode, retries)
		}
		logBackoff(ctx, "network backoff", req, retries, time.Since(start), false, err)
		return nil, nil
	}
	if opts.isBackoffCode(resp.StatusCode) {
		resp.Body.Close()
		if done, _ := backoff.Wait(ctx, resp); done {
			logBackoff(ctx, "application backoff giving up", req, retries, time.Since(start), true, err)
			return nil, handleError(nil, resp.Status, resp.StatusCode, opts.isSuccessCode, retries)
		}
		logBackoff(ctx, "application backoff", req, retries, time.Since(start), false, err)
		return nil, nil
	}
	return resp, nil
}

// maxErrorBody is the most that is read of the body of an unsuccessful response.
const maxErrorBody = 64 << 10

func handleErrorResponse[T any](resp *http.Response, steps int, isOk func(int) bool) (T, *http.Response, []byte, error) {
	var result T
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
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
