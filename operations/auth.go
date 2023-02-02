// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"context"
	"net/http"
)

// Auth represents an authorization mechanism.
type Auth interface {
	// WithAuthorization adds an authorization header or other required
	// authorization information to the provided http.Request.
	WithAuthorization(context.Context, *http.Request) error
}
