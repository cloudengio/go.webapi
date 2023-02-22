// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"context"
)

// Crawled represents a single crawled/downloaded object.
type Crawled[EndpointT any] struct {
	Object  EndpointT
	Content []byte
	Error   error
}

type Fetcher[ScanT, T any] interface {
	Fetch(context.Context, ScanT, chan<- Crawled[T]) error
}

// Crawler is a generic crawler that can be used to iterate over
// a paginated API endpoint (using a Scanner) that enumerates
// objects that can be downloaded using a Fetcher. The Fetcher
// is responsible for decoding the results of each response from
// the paginated API and downloading the objects.
type Crawler[ScanT any, EndpointT any] struct {
	scanner *Scanner[ScanT]
	fetcher Fetcher[ScanT, EndpointT]
}

// NewCrawler creates a new crawler that scans the API using
// the provided Scanner with the result of each scan being passed
// to the Fetcher to download each item returned by the scan.
func NewCrawler[ScanT, EndpointT any](scanner *Scanner[ScanT], fetcher Fetcher[ScanT, EndpointT]) *Crawler[ScanT, EndpointT] {
	return &Crawler[ScanT, EndpointT]{
		scanner: scanner,
		fetcher: fetcher,
	}
}

// Run runs the crawler. It consists of a scan loop that calls
// a Fetcher for each scan response.
func (c *Crawler[ScanT, EndpointT]) Run(ctx context.Context, ch chan<- Crawled[EndpointT]) error {
	defer close(ch)
	for c.scanner.Scan(ctx) {
		resp := c.scanner.Response()
		if err := c.fetcher.Fetch(ctx, resp, ch); err != nil {
			return err
		}
	}
	return c.scanner.Err()
}
