// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersapp

import (
	"context"
	"net/http"
	"time"

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/papersapp/papersappsdk"
)

// APIToken is an implementation of operations.Authorizer for
// a papersapp API token.
type APIToken struct {
	key        string
	token      string
	issued     time.Time
	refreshURL string
}

func NewAPIToken(apiKey, refreshURL string) *APIToken {
	return &APIToken{
		refreshURL: refreshURL,
		key:        apiKey,
	}
}

func (pbt *APIToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	token, err := pbt.Refresh(ctx)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return nil
}

func (pbt *APIToken) Refresh(ctx context.Context) (string, error) {
	if time.Since(pbt.issued) < 50*time.Minute {
		return pbt.token, nil
	}
	ep := operations.NewEndpoint[papersappsdk.Token]()
	req, err := http.NewRequest("GET", pbt.refreshURL+"?api_key="+pbt.key, nil)
	if err != nil {
		return "", err
	}
	tok, _, _, _, err := ep.GetUsingRequest(ctx, req)
	if err != nil {
		return "", err
	}
	pbt.token = tok.Token
	return pbt.token, nil
}
