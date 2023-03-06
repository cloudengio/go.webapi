// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"context"
	"encoding/gob"

	"cloudeng.io/file/content"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

// NewProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func NewProtocolCrawler(
	ctx context.Context, checkpoint Checkpoint,
	fopts FetcherOptions, popts PaginatorOptions, opts ...operations.Option) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error) {
	fetcher, err := NewFetcher(fopts, opts...)
	if err != nil {
		return nil, err
	}
	paginator, err := NewPaginator(ctx, checkpoint, popts)
	if err != nil {
		return nil, err
	}
	scanner := operations.NewScanner(paginator, opts...)
	crawler := operations.NewCrawler(scanner, fetcher)
	return crawler, nil
}

func init() {
	gob.Register(&content.Object[protocolsiosdk.ProtocolPayload, operations.Response]{})
}
