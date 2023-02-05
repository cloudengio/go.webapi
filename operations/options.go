// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"time"
)

// Option represents an option that can be used when creating
// new Endpoints and Streams.
type Option func(o *options)

type options struct {
	rateDelay         time.Duration
	backoffErr        error
	backoffStart      time.Duration
	backoffSteps      int
	backoffStatusCode int
	auth              Auth
	unmarshal         Unmarshal
}

// WithRequestsPerMinute sets the rate for API requests. If not
// specified new requests will be initiated without any delay.
func WithRequestsPerMinute(rpm int) Option {
	return func(o *options) {
		o.rateDelay = time.Minute / time.Duration(rpm)
	}
}

// WithBackoffParameters enables an exponential backoff algorithm that
// is triggered when the request fails in a way that is retryable. A
// request is deemed retryable if the http status code matches
// retryHTTPStatus or if the error returned for the request matches retryErr
// using errors.Is(err, retryErr). retryErr may be nil in which case
// it is essentially ignored.
// First defines the first backoff delay, which is then doubled for every
// consecutive matching error until the download either succeeds or the
// specified number of steps (attempted downloads) is exceeded
// (the download is then deemed to have failed).
func WithBackoffParameters(retryErr error, retryHTTPStatus int, first time.Duration, steps int) Option {
	return func(o *options) {
		o.backoffErr = retryErr
		o.backoffStart = first
		o.backoffSteps = steps
		o.backoffStatusCode = retryHTTPStatus
	}
}

// WithAuth specifies the instance of Auth to use when making requests.
func WithAuth(a Auth) Option {
	return func(o *options) {
		o.auth = a
	}
}

// Unmarshal represents a function that can be used to unmarshal a response
// body.
type Unmarshal func([]byte, any) error

// WithUnmarshal specifies a custom unmarshaling function to use for decoding
// response bodies. The default is json.Unmarshal.
func WithUnmarshal(u Unmarshal) Option {
	return func(o *options) {
		o.unmarshal = u
	}
}
