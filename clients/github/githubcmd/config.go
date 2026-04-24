// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package githubcmd provides support for building command line tools that
// access the GitHub Actions API.
package githubcmd

import (
	"cloudeng.io/webapi/clients/github"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

// Service represents the GitHub-specific configuration for API access.
type Service struct {
	Owner   string `yaml:"owner" cmd:"repository owner (user or organization)"`
	Repo    string `yaml:"repo" cmd:"repository name"`
	PerPage int    `yaml:"per_page" cmd:"number of results per page (max 100, default 30)"`
}

// OptionsForEndpoint returns the operations.Option slice for making API
// requests with the auth and rate-control settings from cfg.
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error) {
	opts := []operations.Option{}
	if len(cfg.KeyID) > 0 {
		opts = append(opts, operations.WithAuth(github.BearerToken{KeyID: cfg.KeyID}))
	}
	rc, err := cfg.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, cfg.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
