// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package benchling provides support for crawling and indexing benchling.com.
package benchling

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

// APIToken is an implementation of operations.Authorizer for
// a benchling API token.
type APIToken struct {
	Token string
}

func (pbt APIToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	token := pbt.Token + ":"
	b64 := base64.RawURLEncoding.EncodeToString([]byte(token))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", b64))
	return nil
}
