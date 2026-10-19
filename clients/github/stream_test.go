// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package github_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudeng.io/webapi/clients/github"
)

func TestStreamWebhookEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("stream") != "1" || r.URL.Query().Get("x") != "y" {
			http.Error(w, "not a stream request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"headers":{"X-Github-Event":["workflow_job"]},"payload":{"action":"queued"}}`)
		fmt.Fprintln(w, `{"payload":{"zen":"x"}}`)
	}))
	defer srv.Close()

	s, err := github.StreamWebhookEvents(context.Background(), srv.URL+"/wait?x=y")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, want := range []struct{ event, payload string }{
		{"workflow_job", `{"action":"queued"}`},
		{"", `{"zen":"x"}`},
	} {
		event, payload, err := s.Next()
		if err != nil {
			t.Fatal(err)
		}
		if event != want.event || string(payload) != want.payload {
			t.Errorf("got %q %s, want %q %s", event, payload, want.event, want.payload)
		}
	}
	if _, _, err := s.Next(); !errors.Is(err, github.ErrRelayClosed) {
		t.Errorf("got %v, want %v", err, github.ErrRelayClosed)
	}
}

func TestStreamWebhookEventsRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no", http.StatusBadRequest)
	}))
	defer srv.Close()
	if _, err := github.StreamWebhookEvents(context.Background(), srv.URL); err == nil {
		t.Error("expected an error")
	}
}

func TestStreamWebhookEventsErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
	}{
		{"not json", "this is not json\n"},
		{"truncated", `{"payload":{"a":`},
		{"too long", `{"payload":"` + strings.Repeat("x", 27<<20) + "\"}\n"},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, tc.body)
		}))
		s, err := github.StreamWebhookEvents(context.Background(), srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = s.Next()
		if err == nil || errors.Is(err, github.ErrRelayClosed) {
			t.Errorf("%s: got %v, want an error other than a clean close", tc.name, err)
		}
		s.Close()
		srv.Close()
	}
}
