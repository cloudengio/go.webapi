// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"context"
	"net/http"
	"time"

	"cloudeng.io/file/content"
)

// Response contains metadata for the result of an operation and is used for
// the response field of the content.Object returned by the Fetcher.
type Response struct {
	// The raw bytes of the response.
	Bytes []byte
	// The encoding used for Bytes.
	Encoding Encoding
	// When the response was received.
	When time.Time

	// Fields copied from the http.Response.
	Headers                http.Header
	Trailers               http.Header
	ContentLength          int64
	StatusCode             int
	ProtoMajor, ProtoMinir int
	TransferEncoding       []string

	// Any error encountered during the operation.
	Error error

	// Checkpoint is an opaque value that can be used to resume an
	// operation at a later time. This is used generally used by implementations
	// of Crawler/Fetcher.
	Checkpoint []byte

	// Current and Total, if non-zero, provide an indication of progress.
	Current int64
	Total   int64
}

func (r *Response) FromHTTPResponse(hr *http.Response) {
	r.Headers = hr.Header
	r.Trailers = hr.Trailer
	r.ContentLength = hr.ContentLength
	r.StatusCode = hr.StatusCode
	r.ProtoMajor = hr.ProtoMajor
	r.ProtoMinir = hr.ProtoMinor
	r.TransferEncoding = hr.TransferEncoding
}

// Fetcher is a generic interface for fetching objects from an API endpoint
// as part of a crawl. The Fetcher extracts/decodes items to be fetched from
// the result of a Scan and then downloads each item.
type Fetcher[ScanT, EndpointT any] interface {
	Fetch(context.Context, ScanT, chan<- []content.Object[EndpointT, Response]) error
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
func (c *Crawler[ScanT, EndpointT]) Run(ctx context.Context, ch chan<- []content.Object[EndpointT, Response]) error {
	defer close(ch)
	i := 0
	for c.scanner.Scan(ctx) {
		resp := c.scanner.Response()
		if err := c.fetcher.Fetch(ctx, resp, ch); err != nil {
			return err
		}
		i++
	}
	return c.scanner.Err()
}

type CrawlHandler[EndpointT any] func(context.Context, content.Object[EndpointT, Response]) error

// RunCrawl is a convenience function that runs a crawler and calls the supplied
// handler for each Object crawled.
func RunCrawl[ScanT, EndpointT any](ctx context.Context, crawler *Crawler[ScanT, EndpointT], handler CrawlHandler[EndpointT]) error {

	errCh := make(chan error)
	ch := make(chan []content.Object[EndpointT, Response])
	go func() {
		errCh <- crawler.Run(ctx, ch)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case objs, ok := <-ch:
			if !ok {
				return <-errCh
			}
			for _, obj := range objs {
				if err := handler(ctx, obj); err != nil {
					return err
				}
			}
		}
	}
}

type DownloadHandler[EndpointT any] func(ctx context.Context, path string, obj content.Object[EndpointT, Response]) error

/*
func ScanDownloaded[EndpointT any](ctx context.Context, handler CrawlHandler[EndpointT]) error {

	return nil
}*/
