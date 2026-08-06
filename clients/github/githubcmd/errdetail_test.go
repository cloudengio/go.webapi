// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package githubcmd

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

// TestErrDetailDefaults verifies that a detail that was never recorded, as
// happens when iteration completes without error, still reports a non-nil body
// and request.
func TestErrDetailDefaults(t *testing.T) {
	var d errDetail
	body, req, err := d.get()
	if body == nil || len(body) != 0 {
		t.Errorf("body: got %v, want empty and non-nil", body)
	}
	if req == nil || req.URL == nil {
		t.Errorf("req: got %v, want non-nil with a non-nil URL", req)
	}
	if err != nil {
		t.Errorf("err: got %v, want nil", err)
	}
}

// TestErrDetailNilRecorded verifies the path taken when no scanner is ever
// created, ie. OptionsForEndpoint fails and the error is recorded with neither
// a body nor a request: the caller still sees the non-nil body and request that
// operations.Scanner.ErrDetail guarantees.
func TestErrDetailNilRecorded(t *testing.T) {
	optsErr := errors.New("bad endpoint options")
	var d errDetail
	d.set(nil, nil, optsErr)
	body, req, err := d.get()
	if body == nil || len(body) != 0 {
		t.Errorf("body: got %v, want empty and non-nil", body)
	}
	if req == nil || req.URL == nil {
		t.Errorf("req: got %v, want non-nil with a non-nil URL", req)
	}
	if !errors.Is(err, optsErr) {
		t.Errorf("err: got %v, want %v", err, optsErr)
	}
}

// TestErrDetailConcurrent exercises the case the mutex exists for: a caller
// consuming the iterator in one goroutine, which writes the detail, while
// another reads it. Run under -race to be meaningful.
func TestErrDetailConcurrent(t *testing.T) {
	const iterations = 1000
	want := errors.New("scan failed")
	d := &errDetail{}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range iterations {
			d.set([]byte("body"), &http.Request{URL: &url.URL{Path: "/p"}}, want)
		}
	}()
	go func() {
		defer wg.Done()
		for range iterations {
			// The values read mid-iteration are not specified; only that
			// reading them races with nothing.
			body, req, _ := d.get()
			if body == nil || req == nil {
				t.Error("get returned a nil body or request")
				return
			}
		}
	}()
	wg.Wait()

	body, req, err := d.get()
	if got, want := string(body), "body"; got != want {
		t.Errorf("body: got %v, want %v", got, want)
	}
	if got, want := req.URL.Path, "/p"; got != want {
		t.Errorf("req path: got %v, want %v", got, want)
	}
	if !errors.Is(err, want) {
		t.Errorf("err: got %v, want %v", err, want)
	}
}
