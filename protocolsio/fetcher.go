// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloudeng.io/file/content"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

const ContentType = "protocols.io/protocol"

type fetcher struct {
	FetcherOptions
	url string
	ep  *operations.Endpoint[protocolsiosdk.ProtocolPayload]
}

func (f *fetcher) fetch(ctx context.Context, p protocolsiosdk.Protocol) (protocolsiosdk.ProtocolPayload, operations.Response) {
	var response operations.Response
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%v", f.url, p.ID), nil)
	if err != nil {
		response.Error = content.Error(err)
		return protocolsiosdk.ProtocolPayload{}, response
	}
	payload, body, enc, resp, err := f.ep.GetUsingRequest(ctx, req)
	if err != nil {
		response.Error = content.Error(err)
		return protocolsiosdk.ProtocolPayload{}, response
	}
	response.Encoding = enc
	response.When = time.Now().Truncate(0)
	response.Bytes = body
	response.FromHTTPResponse(resp)
	return payload, response
}

func (f *fetcher) Fetch(ctx context.Context, page protocolsiosdk.ListProtocolsV3, ch chan<- []content.Object[protocolsiosdk.ProtocolPayload, operations.Response]) error {
	all := []content.Object[protocolsiosdk.ProtocolPayload, operations.Response]{}
	for i, item := range page.Items {
		crawled := content.Object[protocolsiosdk.ProtocolPayload, operations.Response]{
			Type: ContentType,
		}
		var p protocolsiosdk.Protocol
		var outdated bool
		if err := json.Unmarshal(item, &p); err != nil {
			crawled.Response.Error = content.Error(err)
		} else {
			ver, ok := f.VersionMap[p.ID]
			outdated = !ok || ver < p.VersionID
			if outdated {
				crawled.Value, crawled.Response = f.fetch(ctx, p)
			}
		}
		// Always send an object, even if it's empty for a protocol
		// that hasn't changed since we last crawled.
		if i == len(page.Items)-1 {
			cp := Checkpoint{
				CompletedPage: page.Pagination.CurrentPage,
				CurrentPage:   page.Pagination.CurrentPage + 1,
				TotalPages:    page.Pagination.TotalPages,
			}
			buf, _ := json.Marshal(cp)
			crawled.Response.Checkpoint = buf
		}
		crawled.Response.Current = page.Pagination.CurrentPage
		crawled.Response.Total = page.Pagination.TotalPages
		all = append(all, crawled)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- all:
	}
	return nil
}

// PublicBearerToken is an implementation of operations.Authorizer for
// a protocols.io public bearer token.
type PublicBearerToken struct {
	Token string
}

func (pbt PublicBearerToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	req.Header.Add("Bearer", pbt.Token)
	return nil
}

// FetcherOptions represent the options for creating a new Fetcher, including
// a VersionMap which allows for incremental downloading of Protocol objects.
// The VersionMap contains the version ID of a previously downloaded instance
// of that protocol, keyed by it's ID. The fetcher will only redownload a
// protocol object if its version ID has channged. The VersionMap is typically
// built by scanning all previously downloaded protocol objects.
type FetcherOptions struct {
	EndpointURL string
	VersionMap  map[int64]int
}

// NewFetcher returns an instance of operations.Fetcher for protocols.io
// 'GetList' operation.
func NewFetcher(fetcherOpts FetcherOptions, opts ...operations.Option) (operations.Fetcher[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error) {
	return &fetcher{
		FetcherOptions: fetcherOpts,
		url:            fetcherOpts.EndpointURL,
		ep:             operations.NewEndpoint[protocolsiosdk.ProtocolPayload](opts...),
	}, nil
}
