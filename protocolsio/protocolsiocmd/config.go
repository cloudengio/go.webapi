// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsio provides support for working with the
// protocols.io API. It currently provides the abilty to
// crawl public protocols.
package protocolsiocmd

import (
	"context"
	"fmt"
	"strings"

	"cloudeng.io/file/checkpoint"
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
	Endpoints struct {
		ListProtocolsV3 string `yaml:"list_protocols_v3"`
		GetProtocolV4   string `yaml:"get_protocol_v4"`
	}
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:"cache"`
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
	fmt.Fprintf(&out, "endpoint: list v3: %s", c.Endpoints.ListProtocolsV3)
	fmt.Fprintf(&out, "endpoint: get v4: %s", c.Endpoints.GetProtocolV4)
	return out.String()
}

func (c Config) NewPaginator(ctx context.Context) (operations.Paginator[protocolsiosdk.ListProtocolsV3], error) {
	var operation checkpoint.Operation // TODO: checkpoint
	return protocolsio.NewPaginator(ctx, c.Endpoints.ListProtocolsV3, operation)
}

// TODO: load from checkpoint.
func (c Config) NewScanner(ctx context.Context) (*operations.Scanner[protocolsiosdk.ListProtocolsV3], error) {
	paginator, err := c.NewPaginator(ctx)
	if err != nil {
		return nil, err
	}
	return operations.NewScanner(paginator), nil
}

// NewProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func (c Config) NewProtocolCrawler(ctx context.Context) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.Protocol], error) {
	fetcher, err := c.NewFetcher()
	if err != nil {
		return nil, err
	}
	scanner, err := c.NewScanner(ctx)
	if err != nil {
		return nil, err
	}
	crawler := operations.NewCrawler(scanner, fetcher)
	return crawler, nil
}

func (c Config) NewFetcher() (operations.Fetcher[protocolsiosdk.ListProtocolsV3, protocolsiosdk.Protocol], error) {
	opts := []operations.Option{}
	if len(c.Auth.PublicToken) > 0 {
		opts = append(opts,
			operations.WithAuth(protocolsio.PublicBearerToken{Token: c.Auth.PublicToken}))
	}
	rc, err := c.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc))
	return protocolsio.NewFetcher(c.Endpoints.GetProtocolV4, opts...)
}
