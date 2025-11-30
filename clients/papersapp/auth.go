// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersapp

import (
	"context"
	"net/http"
	"time"

	"cloudeng.io/webapi/clients/papersapp/papersappsdk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apitokens"
)

// APIToken is an implementation of operations.Authorizer for
// a papersapp API token.
type APIToken struct {
	refreshKeyID string
	token        papersappsdk.Token
	issued       time.Time
	refreshURL   string
}

func NewAPIToken(refreshKeyID, refreshURL string) *APIToken {
	return &APIToken{
		refreshURL:   refreshURL,
		refreshKeyID: refreshKeyID,
	}
}

func (pbt *APIToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	token, err := pbt.Refresh(ctx)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token.Token)
	return nil
}

func (pbt *APIToken) Refresh(ctx context.Context) (papersappsdk.Token, error) {
	if time.Since(pbt.issued) < 50*time.Minute {
		return pbt.token, nil
	}
	refreshToken, ok := apitokens.TokenFromContext(ctx, pbt.refreshKeyID)
	if !ok {
		return papersappsdk.Token{}, apitokens.NewErrNotFound(pbt.refreshKeyID, pbt.refreshURL)
	}
	ep := operations.NewEndpoint[papersappsdk.Token]()
	req, err := http.NewRequest("GET", pbt.refreshURL+"?api_key="+string(refreshToken.Value()), nil)
	if err != nil {
		return papersappsdk.Token{}, err
	}
	tok, _, _, _, err := ep.IssueRequest(ctx, req)
	if err != nil {
		return papersappsdk.Token{}, err
	}
	pbt.token = tok
	pbt.issued = time.Now()
	return pbt.token, nil
}
