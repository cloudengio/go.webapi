// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersappcmd

import (
	"cloudeng.io/webapi/clients/papersapp"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"cloudeng.io/webapi/operations/apitokens"
)

type Service struct {
	ServiceURL        string `yaml:"service_url" cmd:"papersapp service URL, typically https://api.papers.ai"`
	RefreshTokenURL   string `yaml:"refresh_token_url" cmd:"papersapp refresh token URL, typically https://api.papers.ai/oauth/token"`
	ListItemsPageSize int    `yaml:"list_items_page_size" cmd:"number of items in each page of results, typically 50"`
}

func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service], token *apitokens.T) ([]operations.Option, error) {
	opts := []operations.Option{}
	if tv := token.Token(); len(tv) > 0 {
		opts = append(opts, operations.WithAuth(papersapp.NewAPIToken(string(tv), cfg.Service.RefreshTokenURL)))
	}
	rc, err := cfg.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, cfg.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
