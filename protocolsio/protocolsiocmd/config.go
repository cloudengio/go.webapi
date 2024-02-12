// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsio provides support for working with the
// protocols.io API. It currently provides the ability to
// crawl public protocols.
package protocolsiocmd

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"cloudeng.io/file/checkpoint"
	"cloudeng.io/file/content"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"cloudeng.io/webapi/protocolsio"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

// Auth represents the authentication information required to
// access protocols.io.
type Auth struct {
	PublicToken  string `yaml:"public_token" cmd:"token for protocols.io public data, available from https://www.protocols.io/developers"`
	ClientID     string `yaml:"public_clientid" cmd:"client id for protocols.io public data, available from https://www.protocols.io/developers"`
	ClientSecret string `yaml:"public_secret" cmd:"client secret for protocols.io public data, available from https://www.protocols.io/developers"`
}

// Service represents the protocols.io specific confiugaration options.
type Service struct {
	Filter         string `yaml:"filter" cmd:"filter to apply to protocols.io API calls, typically public"`
	OrderField     string `yaml:"order_field" cmd:"field used to order API responses, typically id"`
	OrderDirection string `yaml:"order_direction" cmd:"order direction to apply to protocols.io API calls, typically asc"`
	Incremental    bool   `yaml:"incremental" cmd:"if true, only download new or updated protocols"`
}

// Config represents the configuration information required to
// access and crawl the protocols.io API.
type Config apicrawlcmd.Crawl[Service]

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

func createVersionMap(ctx context.Context, fs operations.FS, downloads string) (map[int64]int, error) {
	vmap := map[int64]int{}
	store := content.NewStore(fs)
	err := filewalk.ContentsOnly(ctx, fs, downloads, func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
		for _, c := range contents {
			var obj content.Object[protocolsiosdk.ProtocolPayload, operations.Response]
			if _, err := obj.Load(ctx, store, prefix, c.Name); err != nil {
				return err
			}
			vmap[obj.Value.Protocol.ID] = obj.Value.Protocol.VersionID
		}
		return nil
	})
	return vmap, err
}

// NewProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func (c Config) NewProtocolCrawler(ctx context.Context, fs operations.FS, downloadsPath string, op checkpoint.Operation, fv *CrawlFlags, auth Auth) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error) {

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
	paginatorOpts.Parameters.Add("filter", c.Service.Filter)
	paginatorOpts.Parameters.Add("order_field", c.Service.OrderField)
	paginatorOpts.Parameters.Add("order_dir", c.Service.OrderDirection)
	paginatorOpts.Parameters.Add("page_size", strconv.FormatInt(int64(fv.PageSize), 10))
	key := fv.Key
	if len(key) == 0 {
		key = "a"
	}
	paginatorOpts.Parameters.Add("key", key)

	// Fetcher options.
	var fetcherOptions protocolsio.FetcherOptions
	fetcherOptions.EndpointURL = protocolsiosdk.GetProtocolV4Endpoint

	if c.Service.Incremental {
		vmap, err := createVersionMap(ctx, fs, downloadsPath)
		if err != nil {
			return nil, err
		}
		fetcherOptions.VersionMap = vmap
	}

	// General endpoint options.
	opts, err := c.OptionsForEndpoint(auth)
	if err != nil {
		return nil, err
	}

	return protocolsio.NewProtocolCrawler(ctx, cp,
		fetcherOptions,
		paginatorOpts, opts...)

}

func (c Config) OptionsForEndpoint(auth Auth) ([]operations.Option, error) {
	opts := []operations.Option{}
	if len(auth.PublicToken) > 0 {
		opts = append(opts,
			operations.WithAuth(protocolsio.PublicBearerToken{Token: auth.PublicToken}))
	}
	rc, err := c.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, c.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
