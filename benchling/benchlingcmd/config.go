// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchlingcmd

import (
	"context"

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

func (c Config) ListUsersConfig(ctx context.Context) *benchlingsdk.ListUsersParams {
	return &benchlingsdk.ListUsersParams{
		Sort:     &c.Service.UsersSort,
		PageSize: &c.Service.UsersPageSize,
	}
}

func (c Config) ListEntriesConfig(ctx context.Context) *benchlingsdk.ListEntriesParams {
	return &benchlingsdk.ListEntriesParams{
		Sort:     &c.Service.EntriesSort,
		PageSize: &c.Service.EntriesPageSize,
	}
}

func (c Config) ListFoldersConfig(ctx context.Context) *benchlingsdk.ListFoldersParams {
	return &benchlingsdk.ListFoldersParams{
		Sort:     &c.Service.FoldersSort,
		PageSize: &c.Service.FoldersPageSize,
	}
}

func (c Config) ListProjectsConfig(ctx context.Context) *benchlingsdk.ListProjectsParams {
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
	rc, err := c.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, c.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
