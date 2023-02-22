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

type publicBearerToken struct {
	token string
}

func (pbt publicBearerToken) WithAuthorization(ctx context.Context, req *http.Request) error {
	req.Header.Add("Bearer", pbt.token)
	return nil
}

func (c Config) NewFetcher() (operations.Fetcher[protocolsiosdk.ListProtocolsV3, protocolsiosdk.Protocol], error) {
	opts := []operations.Option{}
	if len(c.Auth.PublicToken) > 0 {
		opts = append(opts,
			operations.WithAuth(&publicBearerToken{token: c.Auth.PublicToken}))
	}
	rc, err := c.RateControl.NewRateController()
	if err != nil {
		return nil, err
	}
	opts = append(opts, operations.WithRateController(rc))
	ep := operations.NewEndpoint[protocolsiosdk.ProtocolPayload](opts...)
	return &Fetcher{
		url: c.Endpoints.GetProtocolV4,
		ep:  ep,
	}, nil
}
