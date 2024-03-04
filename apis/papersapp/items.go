// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersapp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"cloudeng.io/webapi/apis/papersapp/papersappsdk"
	"cloudeng.io/webapi/operations"
)

const (
	CollectionType = "papersapp.com/collection"
	ItemType       = "papersapp.com/item"
)

// Item represents a single item in a collection and also includes
// the collection to which it belongs.
type Item struct {
	Item       *papersappsdk.Item
	Collection *papersappsdk.Collection
}

type ItemPaginatorOptions struct {
	EndpointURL string
	Parameters  url.Values
}

type itemPaginator struct {
	ItemPaginatorOptions
	downloaded int
}

// Next implements operations.Paginator.
func (pg *itemPaginator) Next(_ context.Context, items papersappsdk.Items, resp *http.Response) (req *http.Request, done bool, err error) {
	// Pagination for uses a scroll_id parameter to request the next
	// page, but relies on the client counting items until the total
	// number of items is reached. Also note that the scroll_id will be
	// repeated if the end of the list is reached but this requires an
	// extra request to be made and hence counting items is preferable.
	if resp != nil {
		pg.downloaded += len(items.Items)
		if int64(pg.downloaded) >= items.Total {
			done = true
		}
		pg.Parameters.Set("scroll_id", items.ScrollID)
	}
	req, err = http.NewRequest("GET", fmt.Sprintf("%v?%v", pg.EndpointURL, pg.Parameters.Encode()), nil)
	return
}

func NewItemPaginator(opts ItemPaginatorOptions) operations.Paginator[papersappsdk.Items] {
	return &itemPaginator{ItemPaginatorOptions: opts}
}
