// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/webapitestutil"
)

type paginator struct {
	mu      sync.Mutex
	url     string
	nextURL string
	saveURL string
}

func (p *paginator) Next(payload webapitestutil.Paginated, resp *http.Response) (*http.Request, bool, error) {
	if resp == nil {
		// first time through, return the url and mark as paginated.
		req, err := http.NewRequest("GET", p.url, nil)
		return req, false, err
	}
	nextURL := fmt.Sprintf(p.url+"?current=%v", payload.Current+1)
	p.mu.Lock()
	p.nextURL = nextURL
	p.mu.Unlock()
	req, err := http.NewRequest("GET", p.nextURL, nil)
	return req, payload.Current == payload.Last, err
}

func (p *paginator) Save() {
	p.mu.Lock()
	p.saveURL = p.nextURL
	p.mu.Unlock()
}

func TestScanner(t *testing.T) {
	ctx := context.Background()
	handler := &webapitestutil.PaginatedHandler{
		Last: 10,
	}
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	paginator := &paginator{url: srv.URL}
	scanner := operations.NewScanner[webapitestutil.Paginated](paginator)
	expected := 0
	for scanner.Scan(ctx) {
		r := scanner.Response()
		if got, want := r.Payload, expected+1; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := r.Current, expected; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		paginator.mu.Lock()
		if expected == 0 {
			if got, want := paginator.saveURL, ""; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		} else {
			if got, want := paginator.saveURL, fmt.Sprintf(paginator.url+"?current=%v", expected); got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		paginator.mu.Unlock()
		expected++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

type errPaginator struct {
	url      string
	failWhen int
	count    int
}

func (p *errPaginator) Next(payload webapitestutil.Paginated, resp *http.Response) (*http.Request, bool, error) {
	if resp == nil {
		if p.failWhen == 0 {
			return nil, false, fmt.Errorf("fail immediately")
		}
		req, err := http.NewRequest("GET", p.url, nil)
		return req, false, err
	}
	if p.count == p.failWhen {
		return nil, false, fmt.Errorf("fail immediately")
	}
	p.count++
	nextURL := fmt.Sprintf(p.url+"?current=%v", payload.Current+1)
	req, err := http.NewRequest("GET", nextURL, nil)
	return req, payload.Current == payload.Last, err
}

func (p *errPaginator) Save() {}

func TestScannerErrorImmediately(t *testing.T) {
	ctx := context.Background()
	handler := &webapitestutil.PaginatedHandler{
		Last: 10,
	}
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	paginator := &errPaginator{url: srv.URL}
	scanner := operations.NewScanner[webapitestutil.Paginated](paginator)
	for scanner.Scan(ctx) {
		t.Error("expected Scan to return false")
	}
	if err := scanner.Err(); err == nil || err.Error() != "fail immediately" {
		t.Errorf("missing or unexpected error: %v", err)
	}
}

func TestScannerErrorAfterN(t *testing.T) {
	ctx := context.Background()
	handler := &webapitestutil.PaginatedHandler{
		Last: 10,
	}
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	paginator := &errPaginator{url: srv.URL, failWhen: 5}
	scanner := operations.NewScanner[webapitestutil.Paginated](paginator)
	count := 0
	for scanner.Scan(ctx) {
		r := scanner.Response()
		if got, want := r.Current, count; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		count++
	}
	if err := scanner.Err(); err == nil || err.Error() != "fail immediately" {
		t.Errorf("missing or unexpected error: %v", err)
	}
	if got, want := count, paginator.failWhen; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
