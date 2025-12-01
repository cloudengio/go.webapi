// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchling

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"cloudeng.io/webapi/operations/apitokens"
)

// APIToken is an implementation of operations.Authorizer for
// a benchling API token.
type APIToken struct {
	TokenID string
}

// WithAuthorization implements operations.Authorizer.
func (pbt APIToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	token, ok := apitokens.TokenFromContext(ctx, pbt.TokenID)
	if !ok {
		return apitokens.NewErrNotFound(pbt.TokenID, "benchling api token")
	}
	defer token.Clear()
	tok := make([]byte, len(token.Value())+1)
	copy(tok, token.Value())
	tok[len(token.Value())] = ':'
	defer apitokens.ClearToken(tok)
	b64 := base64.RawURLEncoding.EncodeToString(tok)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %v", b64))
	return nil
}
