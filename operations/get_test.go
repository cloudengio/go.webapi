// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"testing"
	"time"

	"cloudeng.io/algo/ratecontrol"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/webapitestutil"
)

type example struct {
	Name  string
	Value int
}

func TestEcho(t *testing.T) {
	ctx := context.Background()

	eg := example{"foo", 42}
	handler := webapitestutil.NewEchoHandler(&eg)
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()

	client := operations.NewEndpoint[example]()

	egr, body, enc, err := client.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := egr, eg; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	data, err := json.Marshal(eg)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := body, data; !bytes.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := enc, operations.JSONEncoding; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRequestRate(t *testing.T) {
	ctx := context.Background()
	timestamps := []time.Time{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		timestamps = append(timestamps, time.Now())
		enc := json.NewEncoder(w)
		if err := enc.Encode(len(timestamps)); err != nil {
			t.Fatal(err)
		}
	})

	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	rc := ratecontrol.New(
		ratecontrol.WithRequestsPerTick(time.Millisecond*100, 1))
	client := operations.NewEndpoint[int](operations.WithRateController(rc, http.StatusTooManyRequests))
	nTimestamps := 5
	for i := range nTimestamps {
		n, _, _, err := client.Get(ctx, srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := n, i+1; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	if got, want := len(timestamps), nTimestamps; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	expectedDelay := time.Millisecond * 100
	for i := 0; i < nTimestamps-1; i++ {
		elapsed := timestamps[i+1].Sub(timestamps[i])
		if got, want := elapsed, expectedDelay; got < (want-want/2) || got > (want+want/2) {
			t.Errorf("got %v, want %v..%v", got, (want - want/2), want+want/2)
		}
	}
}

func TestBackoff(t *testing.T) {
	ctx := context.Background()
	numRetries := 2
	handler := webapitestutil.NewRetryHandler(numRetries)
	srv := webapitestutil.NewServer(handler)
	defer srv.Close()

	rc := ratecontrol.New(ratecontrol.WithExponentialBackoff(time.Millisecond, 2, true))
	client := operations.NewEndpoint[int](operations.WithRateController(rc, http.StatusTooManyRequests))
	n, _, _, err := client.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, numRetries; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

type authorizer struct{}

func (a *authorizer) WithAuthorization(_ context.Context, req *http.Request) error {
	req.Header.Add("something", "secret")
	return nil
}

func TestAuth(t *testing.T) {
	ctx := context.Background()
	handler := webapitestutil.NewHeaderEchoHandler()

	srv := webapitestutil.NewServer(handler)
	defer srv.Close()
	client := operations.NewEndpoint[map[string][]string](operations.WithAuth(&authorizer{}))
	headers, _, _, err := client.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string][]string{
		"Accept-Encoding": {"gzip"},
		"User-Agent":      {"Go-http-client/1.1"},
		"Something":       {"secret"},
	}

	if got := headers; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRequestError(t *testing.T) {
	ctx := context.Background()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	srv := webapitestutil.NewServer(handler)
	defer srv.Close()

	client := operations.NewEndpoint[example]()

	_, _, _, err := client.Get(ctx, srv.URL)
	operr := err.(*operations.Error)
	if got, want := operr.StatusCode, http.StatusNotFound; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTimeout(t *testing.T) {
	ctx := context.Background()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	srv := webapitestutil.NewServer(handler)
	defer srv.Close()

	rc := ratecontrol.New(ratecontrol.WithExponentialBackoff(time.Millisecond, 10, false))
	client := operations.NewEndpoint[example](operations.WithRateController(rc, http.StatusTooManyRequests))

	_, _, _, err := client.Get(ctx, srv.URL)
	operr := err.(*operations.Error)
	if got, want := operr.StatusCode, http.StatusTooManyRequests; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := operr.Attempts, 10; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEndpointIssueRequest(t *testing.T) {
	ctx := context.Background()
	eg := example{"issue", 42}
	srv := webapitestutil.NewServer(webapitestutil.NewEchoHandler(&eg))
	defer srv.Close()

	client := operations.NewEndpoint[example]()
	req, err := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, body, enc, resp, err := client.IssueRequest(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil http.Response")
	}
	if !reflect.DeepEqual(got, eg) {
		t.Errorf("got %v, want %v", got, eg)
	}
	data, _ := json.Marshal(eg)
	if !bytes.Equal(body, data) {
		t.Errorf("body: got %s, want %s", body, data)
	}
	if got, want := enc, operations.JSONEncoding; got != want {
		t.Errorf("encoding: got %v, want %v", got, want)
	}
}

type failAuth struct{}

func (a *failAuth) WithAuthorization(_ context.Context, _ *http.Request) error {
	return fmt.Errorf("auth failed")
}

func TestAuthError(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(webapitestutil.NewEchoHandler(&example{}))
	defer srv.Close()

	client := operations.NewEndpoint[example](operations.WithAuth(&failAuth{}))
	_, _, _, err := client.Get(ctx, srv.URL)
	if err == nil {
		t.Fatal("expected error from failing auth")
	}
}

func TestEncodingContentType(t *testing.T) {
	if got, want := operations.JSONEncoding.ContentType(), "application/json"; got != want {
		t.Errorf("JSONEncoding: got %q, want %q", got, want)
	}
	// An unknown Encoding value should fall back to octet-stream.
	unknown := operations.Encoding(99)
	if got, want := unknown.ContentType(), "application/octet-stream"; got != want {
		t.Errorf("unknown encoding: got %q, want %q", got, want)
	}
}

func TestEndpointGetInvalidURL(t *testing.T) {
	ctx := context.Background()
	client := operations.NewEndpoint[example]()
	_, _, _, err := client.Get(ctx, "://invalid-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestGetNetworkError(t *testing.T) {
	ctx := context.Background()
	client := operations.NewEndpoint[example]()
	// Port 1 is reserved and will refuse connections, triggering a non-retryable
	// network error and covering isErrorRetryable / isErrorRetryableAndLog.
	_, _, _, err := client.Get(ctx, "http://127.0.0.1:1/")
	if err == nil {
		t.Fatal("expected network error")
	}
	// Verify Error.Error() works for the non-nil-Err branch.
	operr, ok := err.(*operations.Error)
	if !ok {
		t.Fatalf("expected *operations.Error, got %T", err)
	}
	if operr.Error() == "" {
		t.Error("Error.Error() returned empty string")
	}
}

func TestErrorMethodNilErr(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := operations.NewEndpoint[example]()
	_, _, _, err := client.Get(ctx, srv.URL)
	operr, ok := err.(*operations.Error)
	if !ok {
		t.Fatalf("expected *operations.Error, got %T", err)
	}
	// When Err is nil, Error() returns Status — covers the nil branch in Error().
	if operr.Error() == "" {
		t.Error("Error.Error() returned empty string for nil-Err case")
	}
}

func TestWithHTTPClient(t *testing.T) {
	ctx := context.Background()
	eg := example{"custom-client", 1}
	srv := webapitestutil.NewServer(webapitestutil.NewEchoHandler(&eg))
	defer srv.Close()

	customClient := &http.Client{Timeout: 10 * time.Second}
	client := operations.NewEndpoint[example](operations.WithHTTPClient(customClient))
	got, _, _, err := client.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, eg) {
		t.Errorf("got %v, want %v", got, eg)
	}
}

func TestWithUnmarshal(t *testing.T) {
	ctx := context.Background()
	eg := example{"custom-unmarshal", 2}
	srv := webapitestutil.NewServer(webapitestutil.NewEchoHandler(&eg))
	defer srv.Close()

	customUnmarshal := json.Unmarshal
	client := operations.NewEndpoint[example](
		operations.WithUnmarshal(customUnmarshal, operations.JSONEncoding),
	)
	got, _, enc, err := client.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, eg) {
		t.Errorf("got %v, want %v", got, eg)
	}
	if got, want := enc, operations.JSONEncoding; got != want {
		t.Errorf("encoding: got %v, want %v", got, want)
	}
}

func TestGetWithSuccessCodes(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"Name":"x","Value":1}`))
	}))
	defer srv.Close()

	// GET default accepts only 200; 201 is an error.
	client := operations.NewEndpoint[example]()
	_, _, _, err := client.Get(ctx, srv.URL)
	if err == nil {
		t.Fatal("expected error for 201 without WithSuccessCodes")
	}

	// WithSuccessCodes(201) makes 201 a success.
	client2 := operations.NewEndpoint[example](
		operations.WithSuccessCodes(http.StatusCreated),
	)
	got, _, _, err := client2.Get(ctx, srv.URL)
	if err != nil {
		t.Fatalf("WithSuccessCodes(201): unexpected error: %v", err)
	}
	if want := (example{"x", 1}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	eg := example{"logged", 3}
	srv := webapitestutil.NewServer(webapitestutil.NewEchoHandler(&eg))
	defer srv.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	client := operations.NewEndpoint[example](operations.WithLogger(logger))
	got, _, _, err := client.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, eg) {
		t.Errorf("got %v, want %v", got, eg)
	}
	if buf.Len() == 0 {
		t.Error("expected log output, got none")
	}
}
