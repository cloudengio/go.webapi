// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package biorxiv

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"cloudeng.io/webapi/operations"
)

type paginator struct {
	serviceURL string
	from, to   time.Time
	cursor     int64
}

func (pg *paginator) Next(ctx context.Context, t Response, r *http.Response) (req *http.Request, done bool, err error) {
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
	var cursor int64
	switch v := msg.Cursor.(type) {
	case int:
		cursor = int64(v)
	case string:
		cursor, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return
		}
	}
	if cursor+msg.Count >= msg.Total {
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
