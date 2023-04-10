// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchlingcmd

import (
	"context"

	"cloudeng.io/net/ratecontrol"
	"cloudeng.io/webapi/benchling"
	"cloudeng.io/webapi/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"

	"cloudeng.io/webapi/operations/apicrawlcmd"
)

type Auth struct {
	APIKey string `yaml:"api_key"`
}

type Service struct {
	ServiceURL       string                              `yaml:"service_url"`
	UsersSort        benchlingsdk.ListUsersParamsSort    `yaml:"users_sort"`
	UsersPageSize    int                                 `yaml:"users_page_size"`
	EntriesSort      benchlingsdk.ListEntriesParamsSort  `yaml:"entries_sort"`
	EntriesPageSize  int                                 `yaml:"entries_page_size"`
	FoldersSort      benchlingsdk.ListFoldersParamsSort  `yaml:"folders_sort"`
	FoldersPageSize  int                                 `yaml:"folders_page_size"`
	ProjectsSort     benchlingsdk.ListProjectsParamsSort `yaml:"projects_sort"`
	ProjectsPageSize int                                 `yaml:"projects_page_size"`
}

type Config apicrawlcmd.Crawl[Service]

func (c Config) ListUsersConfig(_ context.Context) *benchlingsdk.ListUsersParams {
	return &benchlingsdk.ListUsersParams{
		Sort:     &c.Service.UsersSort,
		PageSize: &c.Service.UsersPageSize,
	}
}

func (c Config) ListEntriesConfig(_ context.Context) *benchlingsdk.ListEntriesParams {
	return &benchlingsdk.ListEntriesParams{
		Sort:     &c.Service.EntriesSort,
		PageSize: &c.Service.EntriesPageSize,
	}
}

func (c Config) ListFoldersConfig(_ context.Context) *benchlingsdk.ListFoldersParams {
	return &benchlingsdk.ListFoldersParams{
		Sort:     &c.Service.FoldersSort,
		PageSize: &c.Service.FoldersPageSize,
	}
}

func (c Config) ListProjectsConfig(_ context.Context) *benchlingsdk.ListProjectsParams {
	return &benchlingsdk.ListProjectsParams{
		Sort:     &c.Service.ProjectsSort,
		PageSize: &c.Service.ProjectsPageSize,
	}
}

func (c Config) OptionsForEndpoint(auth Auth) ([]operations.Option, error) {
	opts := []operations.Option{}
	if len(auth.APIKey) > 0 {
		opts = append(opts,
			operations.WithAuth(benchling.APIToken{Token: auth.APIKey}))
	}
	rateCfg := c.RateControl
	rcopts := []ratecontrol.Option{}
	if rateCfg.Rate.BytesPerTick > 0 {
		rcopts = append(rcopts, ratecontrol.WithBytesPerTick(rateCfg.Rate.Tick, rateCfg.Rate.BytesPerTick))
	}
	if rateCfg.Rate.RequestsPerTick > 0 {
		rcopts = append(rcopts, ratecontrol.WithRequestsPerTick(rateCfg.Rate.Tick, rateCfg.Rate.RequestsPerTick))
	}
	if rateCfg.ExponentialBackoff.InitialDelay > 0 {
		rcopts = append(rcopts,
			ratecontrol.WithCustomBackoff(
				func() ratecontrol.Backoff {
					return benchling.NewBackoff(rateCfg.ExponentialBackoff.InitialDelay, rateCfg.ExponentialBackoff.Steps)
				}))
	}
	rc := ratecontrol.New(rcopts...)
	opts = append(opts, operations.WithRateController(rc, c.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
