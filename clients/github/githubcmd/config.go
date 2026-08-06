// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package githubcmd provides support for building command line tools that
// access the GitHub Actions API.
package githubcmd

import (
	"fmt"

	"cloudeng.io/webapi/clients/github"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

// Service represents the GitHub-specific configuration for API access.
type Service struct {
	Owner   string `yaml:"owner" doc:"repository owner, the organization or user name"`
	Repo    string `yaml:"repo" doc:"repository name"`
	PerPage int    `yaml:"per_page" doc:"number of results per page (max 100, default 30)"`
}

func (s Service) Validate() error {
	if s.Owner == "" {
		return fmt.Errorf("githubcmd.Service: missing owner: %+v", s)
	}
	if s.Repo == "" {
		return fmt.Errorf("githubcmd.Service: missing repo: %+v", s)
	}
	if s.PerPage < 1 || s.PerPage > 100 {
		return fmt.Errorf("githubcmd.Service: invalid per_page: must be between 1 and 100: %+v", s)
	}
	return nil
}

// OptionsForEndpoint returns the operations.Option slice for making API
// requests with the auth and rate-control settings from cfg.
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error) {
	opts := []operations.Option{}
	if len(cfg.KeyID) > 0 {
		opts = append(opts, operations.WithAuth(github.BearerToken{KeyID: cfg.KeyID, KeyUser: cfg.UserID}))
	}
	rc, err := cfg.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, cfg.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
