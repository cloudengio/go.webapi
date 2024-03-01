// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchlingcmd

import (
	"context"
	"os"
	"strings"

	"cloudeng.io/net/ratecontrol"
	"cloudeng.io/webapi/benchling"
	"cloudeng.io/webapi/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
	"gopkg.in/yaml.v3"

	"cloudeng.io/webapi/operations/apicrawlcmd"
)

type Auth struct {
	APIKey string `yaml:"api_key" cmd:"API key for benchling"`
}

func (a *Auth) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(a); err != nil {
		return err
	}
	if strings.HasPrefix(a.APIKey, "envvar:") {
		a.APIKey = os.ExpandEnv(strings.TrimSpace(strings.TrimPrefix(a.APIKey, "envvar:")))
	}
	return nil
}

type Service struct {
	ServiceURL       string `yaml:"service_url" cmd:"benchling service URL, typically https://altoslabs.benchling.com/api/v2/"`
	UsersPageSize    int    `yaml:"users_page_size" cmd:"number of users in each page of results, typically 50"`
	EntriesPageSize  int    `yaml:"entries_page_size" cmd:"number of entries in each page of results, typically 50"`
	FoldersPageSize  int    `yaml:"folders_page_size" cmd:"number of folders in each page of results, typically 50"`
	ProjectsPageSize int    `yaml:"projects_page_size" cmd:"number of projects in each page of results, typically 50"`
}

type Config apicrawlcmd.Crawl[Service]

var (
	// The sort order is used to enable checkpointing.
	userSortOrder    benchlingsdk.ListUsersParamsSort    = "modifiedAt:asc"
	entrySortOrder   benchlingsdk.ListEntriesParamsSort  = "modifiedAt:asc"
	folderSortOrder  benchlingsdk.ListFoldersParamsSort  = "modifiedAt:asc"
	projectSortOrder benchlingsdk.ListProjectsParamsSort = "modifiedAt:asc"
)

func (c Config) ListUsersConfig(_ context.Context) *benchlingsdk.ListUsersParams {
	return &benchlingsdk.ListUsersParams{
		Sort:     &userSortOrder,
		PageSize: &c.Service.UsersPageSize,
	}
}

func (c Config) ListEntriesConfig(_ context.Context) *benchlingsdk.ListEntriesParams {
	return &benchlingsdk.ListEntriesParams{
		Sort:     &entrySortOrder,
		PageSize: &c.Service.EntriesPageSize,
	}
}

func (c Config) ListFoldersConfig(_ context.Context) *benchlingsdk.ListFoldersParams {
	return &benchlingsdk.ListFoldersParams{
		Sort:     &folderSortOrder,
		PageSize: &c.Service.FoldersPageSize,
	}
}

func (c Config) ListProjectsConfig(_ context.Context) *benchlingsdk.ListProjectsParams {
	return &benchlingsdk.ListProjectsParams{
		Sort:     &projectSortOrder,
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
