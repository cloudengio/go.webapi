// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package webapitestutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// NewEchoHandler returns an http.Handler that echos the json encoded
// value of the supplied value.
func NewEchoHandler(value any) http.Handler {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := w.Write(body)
		if n != len(body) || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
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

func (wr *withRetry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	w.Write(body)
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
		return
	})
}

// NewServer creates a new httptest.Server using the supplied handler.
func NewServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}
