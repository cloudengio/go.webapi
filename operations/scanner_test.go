// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/webapitestutil"
)

type paginator struct {
	mu         sync.Mutex
	url        string
	currentURL string
	nextURL    string
}

func (p *paginator) Next(_ context.Context, payload webapitestutil.Paginated, resp *http.Response) (*http.Request, bool, error) {
	if resp == nil {
		// first time through, return the url and return false to indicate
		// more pages may follow.
		p.currentURL = p.url
		req, err := http.NewRequest("GET", p.url, nil)
		return req, false, err
	}
	p.mu.Lock()
	p.currentURL = p.nextURL
	nextURL := fmt.Sprintf(p.url+"?current=%v", payload.Current+1)
	p.nextURL = nextURL
	p.mu.Unlock()
	req, err := http.NewRequest("GET", p.nextURL, nil)
	if payload.Current == payload.Last {
		return nil, true, nil
	}
	return req, false, err
}

type authToken struct {
	Token string
}

func (pbt authToken) WithAuthorization(_ context.Context, req *http.Request) error {
	req.Header.Add("Bearer", pbt.Token)
	return nil
}

func TestScanner(t *testing.T) {
	ctx := context.Background()
	handler := &webapitestutil.PaginatedHandler{
		Last: 10,
	}
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	paginator := &paginator{url: srv.URL}
	scanner := operations.NewScanner[webapitestutil.Paginated](paginator, operations.WithAuth(&authToken{"token"}))
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
			if got, want := paginator.currentURL, ""; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		} else {
			if got, want := paginator.currentURL, fmt.Sprintf(paginator.url+"?current=%v", expected); got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		if got, want := paginator.nextURL, fmt.Sprintf(paginator.url+"?current=%v", expected+1); got != want {
			t.Errorf("got %v, want %v", got, want)
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

func (p *errPaginator) Next(_ context.Context, payload webapitestutil.Paginated, resp *http.Response) (*http.Request, bool, error) {
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

// errorHandler responds with a fixed status code and body for every request,
// so tests can exercise the error-detail path of Scanner where an HTTP-level
// error carries a response body and originating request.
type errorHandler struct {
	statusCode int
	body       string
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(h.statusCode)
	_, _ = w.Write([]byte(h.body))
}

// TestScannerErrDetail verifies that when scanning fails because of an HTTP
// error, ErrDetail returns the error together with the non-nil response body
// and the request that caused it.
func TestScannerErrDetail(t *testing.T) {
	ctx := context.Background()
	const body = `{"message":"not found"}`
	srv := webapitestutil.NewServer(&errorHandler{statusCode: http.StatusNotFound, body: body})
	defer srv.Close()
	paginator := &paginator{url: srv.URL}
	scanner := operations.NewScanner[webapitestutil.Paginated](paginator)
	for scanner.Scan(ctx) {
		t.Error("expected Scan to return false")
	}

	err := scanner.Err()
	var opErr *operations.Error
	if !errors.As(err, &opErr) {
		t.Fatalf("expected *operations.Error, got %T: %v", err, err)
	}
	if got, want := opErr.StatusCode, http.StatusNotFound; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	derr, dbody, dreq := scanner.ErrDetail()
	if got, want := derr, err; got != want {
		t.Errorf("ErrDetail error: got %v, want %v", got, want)
	}
	if got, want := string(dbody), body; got != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
	if dreq == nil {
		t.Fatal("expected a non-nil request")
	}
	if got, want := dreq.URL.String(), srv.URL; got != want {
		t.Errorf("request URL: got %v, want %v", got, want)
	}
}

// TestScannerErrDetailPaginatorError verifies that when scanning fails because
// the paginator itself returns an error (rather than an HTTP response), ErrDetail
// returns the error together with an empty, but non-nil, body and request.
func TestScannerErrDetailPaginatorError(t *testing.T) {
	ctx := context.Background()
	handler := &webapitestutil.PaginatedHandler{Last: 10}
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	paginator := &errPaginator{url: srv.URL}
	scanner := operations.NewScanner[webapitestutil.Paginated](paginator)
	for scanner.Scan(ctx) {
		t.Error("expected Scan to return false")
	}

	derr, dbody, dreq := scanner.ErrDetail()
	if derr == nil || derr.Error() != "fail immediately" {
		t.Errorf("missing or unexpected error: %v", derr)
	}
	if dbody == nil {
		t.Error("expected a non-nil body")
	}
	if got, want := len(dbody), 0; got != want {
		t.Errorf("body length: got %v, want %v", got, want)
	}
	if dreq == nil {
		t.Fatal("expected a non-nil request")
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
