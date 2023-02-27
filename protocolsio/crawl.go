// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"context"

	"cloudeng.io/file/checkpoint"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

// NewProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func NewProtocolCrawler(
	ctx context.Context, paginatorEndpoint string, operation checkpoint.Operation,
	fetcherEndpoint string, opts ...operations.Option) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.Protocol], error) {
	fetcher, err := NewFetcher(fetcherEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	paginator, err := NewPaginator(ctx, paginatorEndpoint, operation)
	if err != nil {
		return nil, err
	}
	scanner := operations.NewScanner(paginator)
	crawler := operations.NewCrawler(scanner, fetcher)
	return crawler, nil
}
