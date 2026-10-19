// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations_test

import (
	"bufio"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloudeng.io/algo/ratecontrol"
	"cloudeng.io/webapi/operations"
)

func newStreamRequest(ctx context.Context, t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

// TestIssueStream checks that the response is returned before it is complete,
// with its body unread, so that it can be read as it arrives.
func TestIssueStream(t *testing.T) {
	second := make(chan struct{})
	defer func() {
		select {
		case <-second:
		default:
			close(second)
		}
	}()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "one")
		w.(http.Flusher).Flush()
		select {
		case <-second:
		case <-r.Context().Done():
			return
		}
		fmt.Fprintln(w, "two")
	}))
	defer srv.Close()

	ctx := context.Background()
	resp, err := operations.NewStream().IssueStream(ctx, newStreamRequest(ctx, t, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	if !sc.Scan() || sc.Text() != "one" {
		t.Fatalf("got %q, %v", sc.Text(), sc.Err())
	}
	// The server has not finished: the first line was not waiting for the second.
	close(second)
	if !sc.Scan() || sc.Text() != "two" {
		t.Fatalf("got %q, %v", sc.Text(), sc.Err())
	}
}

func TestIssueStreamErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := context.Background()
	resp, err := operations.NewStream().IssueStream(ctx, newStreamRequest(ctx, t, srv.URL))
	var operr *operations.Error
	if !errors.As(err, &operr) || operr.StatusCode != http.StatusNotFound {
		t.Fatalf("got %v, want a %v error", err, http.StatusNotFound)
	}
	if resp == nil {
		t.Fatal("expected the response")
	}
	defer resp.Body.Close()
	b, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		t.Fatalf("expected readable error body, got error: %v", rerr)
	}
	if got, want := strings.TrimSpace(string(b)), "nope"; got != want {
		t.Errorf("got error body %q, want %q", got, want)
	}
}

func TestIssueStreamBackoff(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		fmt.Fprintln(w, "ok")
	}))
	defer srv.Close()

	rc := ratecontrol.New(ratecontrol.WithExponentialBackoff(time.Millisecond, 5, true))
	ctx := context.Background()
	resp, err := operations.NewEndpoint[string](
		operations.WithRateController(rc, http.StatusTooManyRequests)).IssueStream(ctx, newStreamRequest(ctx, t, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := calls.Load(); got != 3 {
		t.Errorf("got %v calls, want 3", got)
	}
	if b, _ := io.ReadAll(resp.Body); string(b) != "ok\n" {
		t.Errorf("got %q", b)
	}
}

func TestIssueStreamSuccessCodesAndAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, r.Header.Get("something"))
	}))
	defer srv.Close()

	ctx := context.Background()
	if _, err := operations.NewEndpoint[string]().IssueStream(ctx, newStreamRequest(ctx, t, srv.URL)); err == nil {
		t.Error("202 is not a success by default")
	}
	resp, err := operations.NewEndpoint[string](
		operations.WithSuccessCodes(http.StatusAccepted),
		operations.WithAuth(&authorizer{})).IssueStream(ctx, newStreamRequest(ctx, t, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if b, _ := io.ReadAll(resp.Body); string(b) != "secret" {
		t.Errorf("got %q, want the authorization to be applied", b)
	}
}

func TestIssueStreamAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the request should not be made")
	}))
	defer srv.Close()
	ctx := context.Background()
	if _, err := operations.NewEndpoint[string](operations.WithAuth(&failAuth{})).IssueStream(ctx, newStreamRequest(ctx, t, srv.URL)); err == nil {
		t.Error("expected an error")
	}
}

// TestIssueStreamCancel checks that cancelling ctx ends the stream, and that
// it is ctx, not just the request's context, that governs it.
func TestIssueStreamCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "one")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// The request is built with a different context.
	resp, err := operations.NewEndpoint[string]().IssueStream(ctx, newStreamRequest(context.Background(), t, srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	if !sc.Scan() {
		t.Fatal(sc.Err())
	}
	cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for sc.Scan() {
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("cancelling ctx did not end the stream")
	}
}

// TestIssueStreamBoundedErrorBody checks that an error response that never ends
// does not block the caller or exhaust its memory.
func TestIssueStreamBoundedErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		chunk := []byte(strings.Repeat("x", 4096))
		for r.Context().Err() == nil {
			if _, err := w.Write(chunk); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() {
		_, err := operations.NewEndpoint[string]().IssueStream(ctx, newStreamRequest(ctx, t, srv.URL))
		errCh <- err
	}()
	select {
	case err := <-errCh:
		if err == nil {
			t.Error("expected an error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the error body is read without limit")
	}
}

func TestIssueStreamPreCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called")
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := operations.NewStream().IssueStream(ctx, newStreamRequest(ctx, t, srv.URL))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestIssueStreamNonRewindableBodyRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	rc := ratecontrol.New(ratecontrol.WithExponentialBackoff(time.Millisecond, 2, true))
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL, strings.NewReader("payload"))
	if err != nil {
		t.Fatal(err)
	}
	req.GetBody = nil // enforce non-rewindable body

	_, err = operations.NewStream(
		operations.WithRateController(rc, http.StatusTooManyRequests)).IssueStream(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "cannot be rewound") {
		t.Fatalf("expected non-rewindable body error, got: %v", err)
	}
}

func TestIssueStreamNetworkError(t *testing.T) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:0", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = operations.NewStream().IssueStream(ctx, req)
	if err == nil {
		t.Fatal("expected network error for invalid port")
	}
}

type streamItem struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

func (item *streamItem) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	type plain streamItem
	return json.UnmarshalDecode(dec, (*plain)(item))
}

func TestTypedStream(t *testing.T) {
	items := []streamItem{
		{ID: 1, Data: "alpha"},
		{ID: 2, Data: "beta"},
		{ID: 3, Data: "gamma"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for _, item := range items {
			b, err := json.Marshal(item)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "%s\n", b)
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	ep := operations.NewStreamEndpoint[*streamItem]()
	stream, err := ep.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	if resp := stream.Response(); resp == nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %v", resp)
	}
	if stream.Decoder() == nil {
		t.Fatal("expected non-nil decoder")
	}

	var results []streamItem
	var item streamItem
	for {
		err := stream.Next(&item)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}
		results = append(results, item)
	}

	if got, want := len(results), len(items); got != want {
		t.Fatalf("got %d items, want %d", got, want)
	}
	for i := range items {
		if results[i] != items[i] {
			t.Errorf("item %d: got %+v, want %+v", i, results[i], items[i])
		}
	}
}

func TestTypedStreamWithJSONTextOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Output JSON with duplicate keys
		fmt.Fprintln(w, `{"id": 10, "data": "first", "data": "second"}`)
	}))
	defer srv.Close()

	ctx := context.Background()
	// With AllowDuplicateNames(true), jsontext decoder should accept duplicate names
	ep := operations.NewStreamEndpoint[*streamItem](
		operations.WithJSONTextOptions(jsontext.AllowDuplicateNames(true)),
	)
	stream, err := ep.Get(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var item streamItem
	if err := stream.Next(&item); err != nil {
		t.Fatalf("expected successful decode with AllowDuplicateNames: %v", err)
	}
	if item.ID != 10 || item.Data != "second" {
		t.Errorf("got %+v, want ID: 10, Data: second", item)
	}
}

func TestTypedStreamErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "stream not found", http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := context.Background()
	ep := operations.NewStreamEndpoint[*streamItem]()
	stream, err := ep.Get(ctx, srv.URL)
	if err == nil {
		t.Fatal("expected error for 404 status")
	}
	if stream != nil {
		t.Fatal("expected nil stream on error")
	}
	var operr *operations.Error
	if !errors.As(err, &operr) || operr.StatusCode != http.StatusNotFound {
		t.Fatalf("got %v, want 404 operations.Error", err)
	}
}
