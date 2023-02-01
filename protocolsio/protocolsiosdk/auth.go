// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsiosdk

import (
	"context"
	"fmt"
	"net/http"
)

type publicBearerToken string

var bearerTokenKey = publicBearerToken("publicToken")

func WithPublicToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, bearerTokenKey, token)
}

func addAuthHeader(ctx context.Context, req *http.Request) error {
	v := ctx.Value(bearerTokenKey).(string)
	if len(v) > 0 {
		req.Header.Add("Bearer", v)
		return nil
	}
	return fmt.Errorf("no authentication information was found in the context.Context")
}
