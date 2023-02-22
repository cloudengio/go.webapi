// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

type paginator struct {
	CurrentPage int
	TotalPages  int
	url         string
}

func (pg paginator) Next(t protocolsiosdk.ListProtocolsV3, r *http.Response) (req *http.Request, done bool, err error) {
	if r == nil {
		req, err = http.NewRequest("GET", pg.url, nil)
		return
	}
	p := t.Pagination
	pg.CurrentPage = int(p.CurrentPage)
	pg.TotalPages = int(p.TotalPages)
	if done = p.Done(); done {
		return
	}
	u, err := url.Parse(p.NextPage)
	if err != nil {
		return
	}
	np := u.Query().Get("page_id")
	if len(np) == 0 {
		err = fmt.Errorf("%v: failed to find page_id parameter in %v: %#v", p.NextPage, u.String(), p)
		return
	}
	npi, err := strconv.Atoi(np)
	if err != nil {
		err = fmt.Errorf("failed to parse %q: %v", np, err)
	}
	req, err = http.NewRequest("GET", fmt.Sprintf("%v?page_id=%v", pg.url, npi), nil)
	return
}

func (p paginator) Save() {}

func (c Config) NewPaginator() (operations.Paginator[protocolsiosdk.ListProtocolsV3], error) {
	return &paginator{url: c.Endpoints.ListProtocolsV3}, nil
}

// TODO: load from checkpoint.
func (c Config) NewScanner() (*operations.Scanner[protocolsiosdk.ListProtocolsV3], error) {
	paginator, err := c.NewPaginator()
	if err != nil {
		return nil, err
	}
	return operations.NewScanner(paginator), nil
}
