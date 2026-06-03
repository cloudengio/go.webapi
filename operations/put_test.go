// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"cloudeng.io/algo/ratecontrol"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/webapitestutil"
)

// bodyEchoHandler reads the request body and writes it back as the response,
// also exposing the received Content-Type and method via response headers for
// inspection.
type bodyEchoHandler struct {
	t *testing.T
}

func (h *bodyEchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.t.Errorf("reading body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Received-Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("X-Received-Method", r.Method)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		h.t.Errorf("writing response: %v", err)
	}
}

func newBodyEchoServer(t *testing.T) *httptest.Server {
	return webapitestutil.NewServer(&bodyEchoHandler{t: t})
}

func TestPutBasic(t *testing.T) {
	ctx := context.Background()
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	req := example{"hello", 99}
	got, body, enc, err := client.Put(ctx, srv.URL, req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, req) {
		t.Errorf("got %v, want %v", got, req)
	}
	data, _ := json.Marshal(req)
	if !reflect.DeepEqual(body, data) {
		t.Errorf("body: got %s, want %s", body, data)
	}
	if got, want := enc, operations.JSONEncoding; got != want {
		t.Errorf("encoding: got %v, want %v", got, want)
	}
}

func TestPostBasic(t *testing.T) {
	ctx := context.Background()
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	req := example{"world", 7}
	got, _, _, err := client.Post(ctx, srv.URL, req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, req) {
		t.Errorf("got %v, want %v", got, req)
	}
}

func TestPutContentTypeHeader(t *testing.T) {
	ctx := context.Background()
	var receivedContentType, receivedMethod string
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, _, _, err := client.Put(ctx, srv.URL, example{"x", 1})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := receivedContentType, "application/json"; got != want {
		t.Errorf("Content-Type: got %q, want %q", got, want)
	}
	if got, want := receivedMethod, "PUT"; got != want {
		t.Errorf("method: got %q, want %q", got, want)
	}
}

func TestPostContentTypeHeader(t *testing.T) {
	ctx := context.Background()
	var receivedMethod string
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, _, _, err := client.Post(ctx, srv.URL, example{"x", 1})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := receivedMethod, "POST"; got != want {
		t.Errorf("method: got %q, want %q", got, want)
	}
}

// emptyStruct has no fields.
type emptyStruct struct{}

func TestPutEmptyRequestType(t *testing.T) {
	ctx := context.Background()
	var receivedBody []byte
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		resp := example{"response", 1}
		body, _ := json.Marshal(resp)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[emptyStruct, example]()
	got, _, _, err := client.Put(ctx, srv.URL, emptyStruct{})
	if err != nil {
		t.Fatal(err)
	}
	// empty struct marshals to {}
	if got, want := string(receivedBody), "{}"; got != want {
		t.Errorf("body sent: got %q, want %q", got, want)
	}
	if want := (example{"response", 1}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPutEmptyResponseType(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		// return empty JSON object
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, emptyStruct]()
	got, _, _, err := client.Put(ctx, srv.URL, example{"x", 1})
	if err != nil {
		t.Fatal(err)
	}
	if got != (emptyStruct{}) {
		t.Errorf("got %v, want empty struct", got)
	}
}

func TestPutBothEmptyTypes(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[emptyStruct, emptyStruct]()
	got, _, _, err := client.Put(ctx, srv.URL, emptyStruct{})
	if err != nil {
		t.Fatal(err)
	}
	if got != (emptyStruct{}) {
		t.Errorf("got %v, want empty struct", got)
	}
}

func TestPutPrimitiveTypes(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[int, int]()
	got, _, _, err := client.Put(ctx, srv.URL, 42)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := got, 42; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPutStringTypes(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[string, string]()
	got, _, _, err := client.Put(ctx, srv.URL, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := got, "hello"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPutErrorResponse(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, _, _, err := client.Put(ctx, srv.URL, example{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	operr, ok := err.(*operations.Error)
	if !ok {
		t.Fatalf("expected *operations.Error, got %T", err)
	}
	if got, want := operr.StatusCode, http.StatusNotFound; got != want {
		t.Errorf("status code: got %v, want %v", got, want)
	}
}

func TestPutEmptyStructErrorResponse(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[emptyStruct, emptyStruct]()
	_, _, _, err := client.Put(ctx, srv.URL, emptyStruct{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	operr, ok := err.(*operations.Error)
	if !ok {
		t.Fatalf("expected *operations.Error, got %T", err)
	}
	if got, want := operr.StatusCode, http.StatusBadRequest; got != want {
		t.Errorf("status code: got %v, want %v", got, want)
	}
}

func TestIssuePutPostRequestPut(t *testing.T) {
	ctx := context.Background()
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	req, err := http.NewRequestWithContext(ctx, "PUT", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := example{"issue", 5}
	got, body, enc, resp, err := client.IssueRequest(ctx, req, data)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil http.Response")
	}
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v, want %v", got, data)
	}
	encoded, _ := json.Marshal(data)
	if !reflect.DeepEqual(body, encoded) {
		t.Errorf("body: got %s, want %s", body, encoded)
	}
	if got, want := enc, operations.JSONEncoding; got != want {
		t.Errorf("encoding: got %v, want %v", got, want)
	}
}

func TestIssuePutPostRequestPost(t *testing.T) {
	ctx := context.Background()
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	req, err := http.NewRequestWithContext(ctx, "POST", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := example{"post-issue", 10}
	got, _, _, _, err := client.IssueRequest(ctx, req, data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, data) {
		t.Errorf("got %v, want %v", got, data)
	}
}

func TestIssuePutPostRequestEmptyTypes(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[emptyStruct, emptyStruct]()
	req, err := http.NewRequestWithContext(ctx, "PUT", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, _, _, _, err := client.IssueRequest(ctx, req, emptyStruct{})
	if err != nil {
		t.Fatal(err)
	}
	if got != (emptyStruct{}) {
		t.Errorf("got %v, want empty struct", got)
	}
}

func TestPutWithAuth(t *testing.T) {
	ctx := context.Background()
	var receivedHeader string
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("something")
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example](operations.WithAuth(&authorizer{}))
	_, _, _, err := client.Put(ctx, srv.URL, example{"auth", 1})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := receivedHeader, "secret"; got != want {
		t.Errorf("auth header: got %q, want %q", got, want)
	}
}

func TestPutCustomMarshaler(t *testing.T) {
	ctx := context.Background()

	// A custom marshaler that reverses the Name field to distinguish it.
	customMarshal := func(v any) ([]byte, error) {
		if e, ok := v.(example); ok {
			runes := []rune(e.Name)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			e.Name = string(runes)
			return json.Marshal(e)
		}
		return json.Marshal(v)
	}

	var receivedBody []byte
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		resp, _ := json.Marshal(example{"ok", 0})
		_, _ = w.Write(resp)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example](
		operations.WithMarshal(customMarshal, operations.JSONEncoding),
	)
	_, _, _, err := client.Put(ctx, srv.URL, example{"hello", 1})
	if err != nil {
		t.Fatal(err)
	}

	var sent example
	if err := json.Unmarshal(receivedBody, &sent); err != nil {
		t.Fatal(err)
	}
	// name should be reversed: "hello" -> "olleh"
	if got, want := sent.Name, "olleh"; got != want {
		t.Errorf("custom marshal name: got %q, want %q", got, want)
	}
}

func TestPutResponseBodyReturnedOnError(t *testing.T) {
	ctx := context.Background()
	errBody := []byte(`{"error":"not found"}`)
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(errBody)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, body, _, err := client.Put(ctx, srv.URL, example{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !reflect.DeepEqual(body, errBody) {
		t.Errorf("error body: got %s, want %s", body, errBody)
	}
}

func TestPutSliceTypes(t *testing.T) {
	ctx := context.Background()
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[[]example, []example]()
	req := []example{{"a", 1}, {"b", 2}}
	got, _, _, err := client.Put(ctx, srv.URL, req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, req) {
		t.Errorf("got %v, want %v", got, req)
	}
}

func TestPutNilSlice(t *testing.T) {
	ctx := context.Background()
	var receivedBody []byte
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[[]example, []example]()
	got, _, _, err := client.Put(ctx, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	// nil slice marshals to "null"
	if got, want := string(receivedBody), "null"; got != want {
		t.Errorf("body sent: got %q, want %q", got, want)
	}
	// response [] decodes to empty (non-nil) slice
	if len(got) != 0 {
		t.Errorf("got %v, want empty/nil slice", got)
	}
}

// trackingBody wraps an io.ReadCloser and records whether Close was called.
type trackingBody struct {
	io.ReadCloser
	mu     sync.Mutex
	closed bool
}

func (tb *trackingBody) Close() error {
	tb.mu.Lock()
	tb.closed = true
	tb.mu.Unlock()
	return tb.ReadCloser.Close()
}

func (tb *trackingBody) isClosed() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.closed
}

// trackingTransport wraps every response body with a trackingBody so tests can
// assert that all bodies were closed after each call.
type trackingTransport struct {
	mu       sync.Mutex
	delegate http.RoundTripper
	bodies   []*trackingBody
}

func (tt *trackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := tt.delegate.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		tb := &trackingBody{ReadCloser: resp.Body}
		tt.mu.Lock()
		tt.bodies = append(tt.bodies, tb)
		tt.mu.Unlock()
		resp.Body = tb
	}
	return resp, err
}

func (tt *trackingTransport) unclosedCount() int {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	n := 0
	for _, b := range tt.bodies {
		if !b.isClosed() {
			n++
		}
	}
	return n
}

// installTrackingTransport replaces http.DefaultTransport for the duration of
// the test and restores it via t.Cleanup.  Tests that use this must not call
// t.Parallel() because they share global state.
func installTrackingTransport(t *testing.T) *trackingTransport {
	t.Helper()
	orig := http.DefaultTransport
	tt := &trackingTransport{delegate: orig}
	http.DefaultTransport = tt
	t.Cleanup(func() { http.DefaultTransport = orig })
	return tt
}

func TestPutBodyClosedOnSuccess(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, _, _, err := client.Put(ctx, srv.URL, example{"x", 1})
	if err != nil {
		t.Fatal(err)
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after successful PUT", n)
	}
}

func TestPostBodyClosedOnSuccess(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, _, _, err := client.Post(ctx, srv.URL, example{"x", 1})
	if err != nil {
		t.Fatal(err)
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after successful POST", n)
	}
}

func TestPutBodyClosedOnErrorStatus(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	_, _, _, err := client.Put(ctx, srv.URL, example{})
	if err == nil {
		t.Fatal("expected error")
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after error-status PUT", n)
	}
}

func TestPutBodyClosedOnEmptyStructSuccess(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[emptyStruct, emptyStruct]()
	_, _, _, err := client.Put(ctx, srv.URL, emptyStruct{})
	if err != nil {
		t.Fatal(err)
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after empty-struct PUT", n)
	}
}

func TestPutBodyClosedOnBackoffThenSuccess(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	numRetries := 2
	srv := webapitestutil.NewServer(webapitestutil.NewRetryHandler(numRetries))
	defer srv.Close()

	rc := ratecontrol.New(ratecontrol.WithExponentialBackoff(time.Millisecond, numRetries, true))
	client := operations.NewPutEndpoint[int, int](operations.WithRateController(rc, http.StatusTooManyRequests))
	got, _, _, err := client.Put(ctx, srv.URL, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != numRetries {
		t.Errorf("got %v, want %v", got, numRetries)
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after backoff+success PUT (%d total responses)", n, len(tr.bodies))
	}
}

func TestPutBodyClosedOnBackoffGivingUp(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	rc := ratecontrol.New(ratecontrol.WithExponentialBackoff(time.Millisecond, 3, false))
	client := operations.NewPutEndpoint[example, example](operations.WithRateController(rc, http.StatusTooManyRequests))
	_, _, _, err := client.Put(ctx, srv.URL, example{})
	if err == nil {
		t.Fatal("expected error when backoff exhausted")
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after backoff giving up (%d total responses)", n, len(tr.bodies))
	}
}

func TestIssuePutPostBodyClosedOnSuccess(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := newBodyEchoServer(t)
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	req, err := http.NewRequestWithContext(ctx, "PUT", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, _, err = client.IssueRequest(ctx, req, example{"x", 1})
	if err != nil {
		t.Fatal(err)
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after IssuePutPostRequest success", n)
	}
}

func TestIssuePutPostBodyClosedOnErrorStatus(t *testing.T) {
	ctx := context.Background()
	tr := installTrackingTransport(t)
	srv := webapitestutil.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client := operations.NewPutEndpoint[example, example]()
	req, err := http.NewRequestWithContext(ctx, "PUT", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, _, err = client.IssueRequest(ctx, req, example{})
	if err == nil {
		t.Fatal("expected error")
	}
	if n := tr.unclosedCount(); n != 0 {
		t.Errorf("%d response body/bodies not closed after IssuePutPostRequest error status", n)
	}
}
