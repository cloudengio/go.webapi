// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package benchling provides support for crawling and indexing benchling.com.
package benchling

import (
	"context"
	"fmt"
	"net/http"

	"cloudeng.io/webapi/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
)

type tokenIfc interface {
	Next() string
	SetNext(t *string)
}

func getNextToken(p any) *string {
	switch c := p.(type) {
	case Users:
		return c.NextToken
	case Entries:
		return c.NextToken
	case Folders:
		return c.NextToken
	case Projects:
		return c.NextToken
	default:
		panic(fmt.Errorf("unknown type: %T", p))
	}
}

func setNextToken(p any, nextToken *string) {
	switch c := p.(type) {
	case *benchlingsdk.ListEntriesParams:
		c.NextToken = nextToken
	case *benchlingsdk.ListUsersParams:
		c.NextToken = nextToken
	case *benchlingsdk.ListFoldersParams:
		c.NextToken = nextToken
	case *benchlingsdk.ListProjectsParams:
		c.NextToken = nextToken
	default:
		panic(fmt.Errorf("unknown type: %T", p))
	}
}

func createRequest(serviceURL string, params any) (*http.Request, error) {
	switch c := params.(type) {
	case *benchlingsdk.ListEntriesParams:
		return benchlingsdk.NewListEntriesRequest(serviceURL, c)
	case *benchlingsdk.ListUsersParams:
		return benchlingsdk.NewListUsersRequest(serviceURL, c)
	case *benchlingsdk.ListFoldersParams:
		return benchlingsdk.NewListFoldersRequest(serviceURL, c)
	case *benchlingsdk.ListProjectsParams:
		return benchlingsdk.NewListProjectsRequest(serviceURL, c)
	default:
		panic(fmt.Errorf("unknown type: %T", params))
	}
}

type ScanPayload[T any] struct {
	NextToken *string
	Payload   []T
}

type paginator[T any] struct {
	serviceURL string
	payload    ScanPayload[T]
	params     any
}

func (pg *paginator[T]) Next(ctx context.Context, t T, r *http.Response) (req *http.Request, done bool, err error) {
	if r == nil {
		req, err = createRequest(pg.serviceURL, pg.params)
		return
	}
	nt := getNextToken(t)
	done = nt == nil || len(*nt) == 0
	setNextToken(pg.params, nt)
	req, err = createRequest(pg.serviceURL, pg.params)
	return
}

func newPaginator[T Scanners](ctx context.Context, serviceURL string, params any) operations.Paginator[T] {
	pg := &paginator[T]{
		serviceURL: serviceURL,
		params:     params,
	}
	return pg
}
