// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package webapitestutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
)

// NewEchoHandler returns an http.Handler that echos the json encoded
// value of the supplied value.
func NewEchoHandler(value any) http.Handler {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n, err := w.Write(body)
		if n != len(body) || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// NewRetryHandler returns an http.Handler that returns an http.StatusTooManyRequests
// until retries in reached in which case it returns an http.StatusOK with
// a body containing the json encoded value of retries.
func NewRetryHandler(retries int) http.Handler {
	return &withRetry{retries: retries}
}

type withRetry struct {
	retries int
	count   int
}

func (wr *withRetry) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if wr.count < wr.retries {
		wr.count++
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	body, err := json.Marshal(wr.retries)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// NewHeaderEchoHandler returns an http.Handler that returns the json encoded
// value of the request headers as its response body.
func NewHeaderEchoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		n, err := w.Write(body)
		if n != len(body) || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// Paginated is an example return type for a paginated API.
type Paginated struct {
	Payload int
	Current int
	Last    int
}

// PaginatedHandler is an example http.Handler that returns a Paginated
// response, that is accepts a parameter called 'current' which is
// the page to returned. Last should be initialized the number of available
// pages.
type PaginatedHandler struct {
	invocations int
	Last        int
}

func (ph *PaginatedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.invocations++
	current, err := strconv.ParseInt(r.URL.Query().Get("current"), 10, 32)
	if err != nil {
		current = 0
	}
	if current > int64(ph.Last) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p := Paginated{
		Payload: ph.invocations,
		Current: int(current),
		Last:    ph.Last,
	}
	body, err := json.Marshal(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// NewServer creates a new httptest.Server using the supplied handler.
func NewServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}
