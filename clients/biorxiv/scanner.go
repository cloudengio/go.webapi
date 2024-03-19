// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package biorxiv

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"cloudeng.io/webapi/clients/biorxiv/internal"
	"cloudeng.io/webapi/operations"
)

type paginator struct {
	serviceURL string
	from, to   time.Time
	cursor     int64
}

func (pg *paginator) Next(_ context.Context, t Response, r *http.Response) (req *http.Request, done bool, err error) {
	if r == nil {
		var u string
		u, err = url.JoinPath(pg.serviceURL, pg.from.Format("2006-01-02"), pg.to.Format("2006-01-02"), strconv.FormatInt(pg.cursor, 10))
		if err != nil {
			return nil, false, err
		}
		req, err = http.NewRequest("GET", u, nil)
		return
	}
	if len(t.Messages) == 0 {
		done = true
		return
	}
	msg := t.Messages[0]
	cursor, err := internal.AsInt64(msg.Cursor)
	if err != nil {
		err = fmt.Errorf("unexpected cursor: %v: %v", msg.Cursor, err)
		return
	}

	total, err := internal.AsInt64(msg.Total)
	if err != nil {
		err = fmt.Errorf("unexpected total: %v: %v", msg.Total, err)
		return
	}
	if cursor+msg.Count >= total {
		done = true
		return
	}
	u, err := url.JoinPath(pg.serviceURL, pg.from.Format("2006-01-02"), pg.to.Format("2006-01-02"),
		strconv.FormatInt(cursor+msg.Count, 10))
	if err != nil {
		return
	}
	req, err = http.NewRequest("GET", u, nil)
	return
}

// NewScanner returns an instance of operations.Scanner for scanning
// biorxiv or medrxiv via api.biorxiv.org. The from, to and cursor
// values corresponding to URL path components as documented at:
// https://api.biorxiv.org/
func NewScanner(serviceURL string, from, to time.Time, cursor int64, opts ...operations.Option) *operations.Scanner[Response] {
	pg := &paginator{
		serviceURL: serviceURL,
		from:       from,
		to:         to,
		cursor:     cursor,
	}
	return operations.NewScanner[Response](pg, opts...)
}
