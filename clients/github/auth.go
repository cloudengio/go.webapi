// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package github provides a client for the GitHub REST API, with a focus
// on the Actions endpoints (runs, jobs, runners).
package github

import (
	"context"
	"net/http"

	"cloudeng.io/webapi/operations/apitokens"
)

// BearerToken implements operations.Auth for GitHub personal access tokens
// and GitHub Apps installation tokens. The token is retrieved from the context
// via the apitokens package using the configured KeyID.
type BearerToken struct {
	KeyID string
}

// WithAuthorization implements operations.Auth. It sets the Authorization
// header and the required GitHub API headers on the request.
func (bt BearerToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	token, ok := apitokens.TokenFromContext(ctx, bt.KeyID)
	if !ok {
		return apitokens.NewErrNotFound(bt.KeyID, "github bearer token")
	}
	defer token.Clear()
	req.Header.Set("Authorization", "Bearer "+string(token.Value()))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return nil
}
