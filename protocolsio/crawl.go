// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

// NewProtocolCrawler creates a new instance of operations.Crawler
// that can be used to crawl/download protocols on protocols.io.
func (c Config) NewProtocolCrawler() (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.Protocol], error) {
	fetcher, err := c.NewFetcher()
	if err != nil {
		return nil, err
	}
	scanner, err := c.NewScanner()
	if err != nil {
		return nil, err
	}
	crawler := operations.NewCrawler(scanner, fetcher)
	return crawler, nil
}
