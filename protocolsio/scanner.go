// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"cloudeng.io/file/checkpoint"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

type paginator struct {
	op          checkpoint.Operation
	currentPage int
	totalPages  int
	done        bool
	url         string
}

func (pg paginator) Next(t protocolsiosdk.ListProtocolsV3, r *http.Response) (req *http.Request, done bool, err error) {
	if r == nil {
		req, err = http.NewRequest("GET", pg.url, nil)
		return
	}
	p := t.Pagination
	pg.currentPage = int(p.CurrentPage)
	pg.totalPages = int(p.TotalPages)
	if p.CurrentPage == p.TotalPages {
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

func (pg paginator) Save() {
}

func NewPaginator(ctx context.Context, endpoint string, operation checkpoint.Operation) (operations.Paginator[protocolsiosdk.ListProtocolsV3], error) {
	pg := &paginator{url: endpoint, op: operation}
	// TODO: checkpointing
	return pg, nil
}
