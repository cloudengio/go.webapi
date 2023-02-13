// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"cloudeng.io/net/ratecontrol"
)

// Option represents an option that can be used when creating
// new Endpoints and Streams.
type Option func(o *options)

type options struct {
	backoffStatusCodes []int
	rateController     *ratecontrol.Controller
	auth               Auth
	unmarshal          Unmarshal
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
