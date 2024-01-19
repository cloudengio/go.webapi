// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersappcmd

import (
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"cloudeng.io/webapi/papersapp"
)

type Auth struct {
	APIKey string `yaml:"api_key" cmd:"API key for papersapp"`
}

type Service struct {
	ServiceURL        string `yaml:"service_url" cmd:"papersapp service URL, typically https://api.papers.ai"`
	RefreshTokenURL   string `yaml:"refresh_token_url" cmd:"papersapp refresh token URL, typically https://api.papers.ai/oauth/token"`
	ListItemsPageSize int    `yaml:"list_items_page_size" cmd:"number of items in each page of results, typically 50"`
}

type Config apicrawlcmd.Crawl[Service]

func (c Config) OptionsForEndpoint(auth Auth) ([]operations.Option, error) {
	opts := []operations.Option{}
	if len(auth.APIKey) > 0 {
		opts = append(opts,
			operations.WithAuth(papersapp.NewAPIToken(auth.APIKey, c.Service.RefreshTokenURL)))
	}
	rc, err := c.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, c.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
