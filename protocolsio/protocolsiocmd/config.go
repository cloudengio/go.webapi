// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsio provides support for working with the
// protocols.io API. It currently provides the abilty to
// crawl public protocols.
package protocolsiocmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cloudeng.io/file/checkpoint"
	"cloudeng.io/file/content"
	"cloudeng.io/file/crawl/crawlcmd"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

// Config represents the configuration information required to
// access and crawl the protocols.io API.
type Config struct {
	Auth struct {
		PublicToken  string `yaml:"public_token"`
		ClientID     string `yaml:"public_clientid"`
		ClientSecret string `yaml:"public_secret"`
	}
	RateControl    crawlcmd.RateControl      `yaml:",inline"`
	Cache          crawlcmd.CrawlCacheConfig `yaml:",inline"`
	Filter         string                    `yaml:"filter"`
	OrderField     string                    `yaml:"order_field"`
	OrderDirection string                    `yaml:"order_direction"`
	Incremental    bool                      `yaml:"incremental"`
}

func (c Config) String() string {
	var out strings.Builder
	out.WriteString("protocolsio:\n")
	if len(c.Auth.PublicToken) > 0 {
		fmt.Fprintf(&out, " public token: **redacted**\n")
	}
	if len(c.Auth.ClientID) > 0 {
		fmt.Fprintf(&out, "    client id: %s\n", c.Auth.ClientID)
	}
	if len(c.Auth.ClientSecret) > 0 {
		fmt.Fprintf(&out, "client secret: **redacted**\n")
	}
	return out.String()
}

func latestCheckpoint(ctx context.Context, op checkpoint.Operation) (protocolsio.Checkpoint, error) {
	if op == nil {
		return protocolsio.Checkpoint{}, nil
	}
	buf, err := op.Latest(ctx)
	if err != nil {
		return protocolsio.Checkpoint{}, err
	}
	if len(buf) == 0 {
		return protocolsio.Checkpoint{}, nil
	}
	var cp protocolsio.Checkpoint
	err = json.Unmarshal(buf, &cp)
	return cp, err
}

func createVersionMap(cachePath, checkpointPath string) (map[int64]int, error) {
	vmap := map[int64]int{}
	err := filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if path == checkpointPath {
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		var obj content.Object[protocolsiosdk.ProtocolPayload, operations.Response]
		if err := obj.ReadObjectBinary(path); err != nil {
			return err
		}
		vmap[obj.Value.Protocol.ID] = obj.Value.Protocol.VersionID
		return nil
	})
	return vmap, err
}

// NewProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func (c Config) NewProtocolCrawler(ctx context.Context, op checkpoint.Operation, fv *CrawlFlags) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error) {

	cp, err := latestCheckpoint(ctx, op)
	if err != nil {
		return nil, err
	}

	// Pagination options.
	var paginatorOpts protocolsio.PaginatorOptions
	paginatorOpts.From = fv.Pages.From
	if fv.Pages.To != 0 && !fv.Pages.ExtendsToEnd {
		paginatorOpts.To = fv.Pages.To
	}
	paginatorOpts.EndpointURL = protocolsiosdk.ListProtocolsV3Endpoint
	paginatorOpts.Parameters = url.Values{}
	paginatorOpts.Parameters.Add("filter", c.Filter)
	paginatorOpts.Parameters.Add("order_field", c.OrderField)
	paginatorOpts.Parameters.Add("order_dir", c.OrderDirection)
	paginatorOpts.Parameters.Add("page_size", strconv.FormatInt(int64(fv.PageSize), 10))
	paginatorOpts.Parameters.Add("key", fv.Key)

	// Fetcher options.
	var fetcherOptions protocolsio.FetcherOptions
	fetcherOptions.EndpointURL = protocolsiosdk.GetProtocolV4Endpoint

	if c.Incremental {
		vmap, err := createVersionMap(c.Cache.Prefix, c.Cache.Checkpoint)
		if err != nil {
			return nil, err
		}
		fetcherOptions.VersionMap = vmap
	}

	// General endpoint options.
	opts, err := c.OptionsForEndpoint()
	if err != nil {
		return nil, err
	}

	return protocolsio.NewProtocolCrawler(ctx, cp,
		fetcherOptions,
		paginatorOpts, opts...)

}

func (c Config) OptionsForEndpoint() ([]operations.Option, error) {
	opts := []operations.Option{}
	if len(c.Auth.PublicToken) > 0 {
		opts = append(opts,
			operations.WithAuth(protocolsio.PublicBearerToken{Token: c.Auth.PublicToken}))
	}
	rc, err := c.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, c.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
