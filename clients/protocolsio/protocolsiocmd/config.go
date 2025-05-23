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
	"sync"

	"cloudeng.io/file/checkpoint"
	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/logging/ctxlog"
	"cloudeng.io/webapi/clients/protocolsio"
	"cloudeng.io/webapi/clients/protocolsio/protocolsiosdk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"cloudeng.io/webapi/operations/apitokens"
)

// Service represents the protocols.io specific confiugaration options.
type Service struct {
	Filter         string `yaml:"filter" cmd:"filter to apply to protocols.io API calls, typically public"`
	OrderField     string `yaml:"order_field" cmd:"field used to order API responses, typically id"`
	OrderDirection string `yaml:"order_direction" cmd:"order direction to apply to protocols.io API calls, typically asc"`
	Incremental    bool   `yaml:"incremental" cmd:"if true, only download new or updated protocols"`
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

func createVersionMap(ctx context.Context, fs operations.FS, concurrency int, downloads string) (map[int64]int, error) {
	vmap := map[int64]int{}
	var mu sync.Mutex
	store := stores.New(fs, concurrency)
	err := filewalk.ContentsOnly(ctx, fs, downloads, func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
		if err != nil {
			ctxlog.Error(ctx, "protocols.io: error", "prefix", prefix, "err", err)
		}
		names := make([]string, len(contents))
		for i, c := range contents {
			names[i] = c.Name
		}
		return store.ReadV(ctx, prefix, names, func(_ context.Context, _, _ string, _ content.Type, buf []byte, err error) error {
			if err != nil {
				return err
			}
			var obj content.Object[protocolsiosdk.ProtocolPayload, operations.Response]
			if err := obj.Decode(buf); err != nil {
				return err
			}
			mu.Lock()
			vmap[obj.Value.Protocol.ID] = obj.Value.Protocol.VersionID
			mu.Unlock()
			return nil
		})
	})
	return vmap, err
}

// newProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func newProtocolCrawler(ctx context.Context, cfg apicrawlcmd.Crawl[Service], fs operations.FS, downloadsPath string, op checkpoint.Operation, fv *CrawlFlags, token *apitokens.T) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error) {

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
	paginatorOpts.Parameters.Add("filter", cfg.Service.Filter)
	paginatorOpts.Parameters.Add("order_field", cfg.Service.OrderField)
	paginatorOpts.Parameters.Add("order_dir", cfg.Service.OrderDirection)
	paginatorOpts.Parameters.Add("page_size", strconv.FormatInt(int64(fv.PageSize), 10))
	key := fv.Key
	if len(key) == 0 {
		key = "a"
	}
	paginatorOpts.Parameters.Add("key", key)

	// Fetcher options.
	var fetcherOptions protocolsio.FetcherOptions
	fetcherOptions.EndpointURL = protocolsiosdk.GetProtocolV4Endpoint

	if cfg.Service.Incremental {
		vmap, err := createVersionMap(ctx, fs, cfg.Cache.Concurrency, downloadsPath)
		if err != nil {
			return nil, err
		}
		fetcherOptions.VersionMap = vmap
	}

	// General endpoint options.
	opts, err := OptionsForEndpoint(cfg, token)
	if err != nil {
		return nil, err
	}

	return protocolsio.NewProtocolCrawler(ctx, cp,
		fetcherOptions,
		paginatorOpts, opts...)

}

func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service], token *apitokens.T) ([]operations.Option, error) {
	opts := []operations.Option{}
	if tv := token.Token(); len(tv) > 0 {
		opts = append(opts, operations.WithAuth(protocolsio.PublicBearerToken{Token: string(tv)}))
	}
	rc, err := cfg.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc, cfg.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
