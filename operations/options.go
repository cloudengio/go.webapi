// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"cloudeng.io/algo/ratecontrol"
	"cloudeng.io/logging/ctxlog"
)

// Signer represents a function that can be used to sign requests, e.g. by
// adding appropriate headers. This is used for operations that require signing
// of requests. Signer is called with the payload to be signed and the header
// to which signature information should be added.
type Signer func(header http.Header, payload []byte) error

// Option represents an option that can be used when creating
// new Endpoints and Streams.
type Option func(o *options)

type options struct {
	backoffStatusCodes []int
	rateController     *ratecontrol.Controller
	auth               Auth
	unmarshal          Unmarshal
	encoding           Encoding
	marshal            Marshal
	marshalEncoding    Encoding
	logger             *slog.Logger
	client             *http.Client
	signer             Signer
	successCodes       []int
}

func handleOptions(options *options, putPost bool, opts ...Option) {
	for _, fn := range opts {
		fn(options)
	}
	if options.rateController == nil {
		options.rateController = ratecontrol.New()
	}
	if options.unmarshal == nil {
		options.unmarshal = json.Unmarshal
		options.encoding = JSONEncoding
	}
	if options.marshal == nil {
		options.marshal = json.Marshal
		options.marshalEncoding = JSONEncoding
	}
	if options.logger == nil {
		options.logger = slog.New(slog.DiscardHandler)
	}
	if options.client == nil {
		options.client = http.DefaultClient
	}
	if len(options.successCodes) == 0 {
		if putPost {
			options.successCodes = []int{http.StatusOK, http.StatusAccepted}
		} else {
			options.successCodes = []int{http.StatusOK}
		}
	}
}

// WithRateController sets the rate controller to use to enforce rate
// control and backoff.
func WithRateController(rc *ratecontrol.Controller, statusCodes ...int) Option {
	return func(o *options) {
		o.backoffStatusCodes = statusCodes
		o.rateController = rc
	}
}

// WithAuth specifies the instance of Auth to use when making requests.
func WithAuth(a Auth) Option {
	return func(o *options) {
		o.auth = a
	}
}

// WithLogger specifies the logger to use for logging request and response
// information. If not specified, no logging is performed.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithHTTPClient specifies the http.Client to use for making requests. If not
// specified, http.DefaultClient is used.
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) {
		o.client = client
	}
}

// WithSigner specifies a Signer function to use for signing requests.
func WithSigner(signer Signer) Option {
	return func(o *options) {
		o.signer = signer
	}
}

// WithSuccessCodes specifies the HTTP status codes that should be considered
// successful responses. If not specified, only http.StatusOK (200) is
// considered a successful response for Get operations and http.StatusOK (200)
// http.StatusAccepted or for Put/Post operations.
func WithSuccessCodes(codes ...int) Option {
	return func(o *options) {
		o.successCodes = slices.Clone(codes)
	}
}

// Unmarshal represents a function that can be used to unmarshal a response
// body.
type Unmarshal func([]byte, any) error

// Marshal represents a function that can be used to marshal a request body.
type Marshal func(any) ([]byte, error)

// WithUnmarshal specifies a custom unmarshaling function to use for decoding
// response bodies. The default is json.Unmarshal.
func WithUnmarshal(u Unmarshal, e Encoding) Option {
	return func(o *options) {
		o.unmarshal = u
		o.encoding = e
	}
}

// WithMarshal specifies a custom marshaling function to use for encoding
// request bodies. The default is json.Marshal.
func WithMarshal(marshal Marshal, e Encoding) Option {
	return func(o *options) {
		o.marshal = marshal
		o.marshalEncoding = e
	}
}

func (o options) isBackoffCode(code int) bool {
	return slices.Contains(o.backoffStatusCodes, code)
}

func isErrorRetryable(err error) (string, bool) {
	if errors.Is(err, context.DeadlineExceeded) {
		return "context.DeadlineExceeded", true
	}
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return "timeout", true
	}
	if strings.HasSuffix(err.Error(), ": connection reset by peer") {
		return "connection reset by peer", true
	}
	if strings.Contains(err.Error(), "TLS handshake") {
		return "TLS handshake", true
	}
	return "cannot retry", false
}

func (o options) isErrorRetryableAndLog(ctx context.Context, req *http.Request, err error) bool {
	msg, retryable := isErrorRetryable(err)
	grp := slog.Group("req", "url", req.URL, "err", err, "retryable", retryable)
	ctxlog.Info(ctx, msg, grp)
	return retryable
}

func logBackoff(ctx context.Context, msg string, req *http.Request, retries int, took time.Duration, done bool, err error) {
	grp := slog.Group("req", "url", req.URL, "retries", retries, "took", took, "done", done, "err", err)
	ctxlog.Info(ctx, msg, grp)
}

func (o options) isSuccessCode(code int) bool {
	return slices.Contains(o.successCodes, code)
}
