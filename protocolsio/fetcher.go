// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

type Fetcher struct {
	url string
	ep  *operations.Endpoint[protocolsiosdk.ProtocolPayload]
}

func (f *Fetcher) Fetch(ctx context.Context, page protocolsiosdk.ListProtocolsV3, ch chan<- operations.Crawled[protocolsiosdk.Protocol]) error {
	for _, item := range page.Items {
		crawled := operations.Crawled[protocolsiosdk.Protocol]{}
		var p protocolsiosdk.Protocol
		if err := json.Unmarshal(item, &p); err != nil {
			crawled.Error = err
		} else {
			u := fmt.Sprintf("%s/%v", f.url, p.ID)
			payload, body, err := f.ep.Get(ctx, u)
			crawled.Content = body
			if payload.StatusCode != 0 {
				crawled.Error = fmt.Errorf("status code: %v", payload.StatusCode)
			} else {
				crawled.Object = payload.Protocol
				crawled.Error = err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- crawled:
		}
	}
	return nil
}

type PublicBearerToken struct {
	Token string
}

func (pbt PublicBearerToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	req.Header.Add("Bearer", pbt.Token)
	return nil
}

func NewFetcher(endpoint string, opts ...operations.Option) (operations.Fetcher[protocolsiosdk.ListProtocolsV3, protocolsiosdk.Protocol], error) {
	ep := operations.NewEndpoint[protocolsiosdk.ProtocolPayload](opts...)
	return &Fetcher{
		url: endpoint,
		ep:  ep,
	}, nil
}
