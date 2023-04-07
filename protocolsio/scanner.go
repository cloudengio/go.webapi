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

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

type paginator struct {
	completedPage int64
	currentPage   int64
	totalPages    int64
	PaginatorOptions
}

func (pg *paginator) urlfor(page int64, first bool) string {
	if first && pg.From != 0 {
		page = int64(pg.From)
	}
	pg.Parameters.Set("page_id", strconv.FormatInt(page, 10))
	u := fmt.Sprintf("%v?%v", pg.EndpointURL, pg.Parameters.Encode())
	return u
}

func (pg *paginator) Next(_ context.Context, t protocolsiosdk.ListProtocolsV3, r *http.Response) (req *http.Request, done bool, err error) {
	if r == nil {
		req, err = http.NewRequest("GET", pg.urlfor(pg.currentPage, true), nil)
		return
	}
	p := t.Pagination
	pg.completedPage = p.CurrentPage
	pg.totalPages = p.TotalPages
	if done = (p.CurrentPage >= p.TotalPages) || (pg.To != 0 && pg.completedPage >= int64(pg.To)); done {
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
	npi, err := strconv.ParseInt(np, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse %q: %v", np, err)
		return
	}
	pg.currentPage = npi
	req, err = http.NewRequest("GET", pg.urlfor(npi, false), nil)
	return
}

type PaginatorOptions struct {
	EndpointURL string
	Parameters  url.Values
	From, To    int
}

// NewPaginator returns an instance of operations.Paginator for protocols.io
// 'GetList' operation.
func NewPaginator(_ context.Context, cp Checkpoint, opts PaginatorOptions) (operations.Paginator[protocolsiosdk.ListProtocolsV3], error) {
	pg := &paginator{PaginatorOptions: opts}
	pg.Parameters.Set("fields", "id,version_id")
	pg.completedPage = cp.CompletedPage
	pg.currentPage = cp.CurrentPage
	if pg.completedPage == 0 && pg.currentPage == 0 {
		pg.currentPage = 1
	}
	return pg, nil
}
