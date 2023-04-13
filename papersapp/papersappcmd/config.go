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
	APIKey string `yaml:"api_key"`
}

type Service struct {
	ServiceURL        string `yaml:"service_url"`
	RefreshTokenURL   string `yaml:"refresh_token_url"`
	ListItemsPageSize int    `yaml:"list_items_page_size"`
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
