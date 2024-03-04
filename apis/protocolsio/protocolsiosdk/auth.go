// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsiosdk

import (
	"context"
)

type publicBearerToken string

var bearerTokenKey = publicBearerToken("publicToken")

func WithPublicToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, bearerTokenKey, token)
}
