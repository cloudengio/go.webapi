// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apicrawlcmd_test

import (
	"testing"
	"time"

	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

const apiCrawlSpec = `
api1:
  cache_checkpoint: path/to/checkpoint
  cache_sharding_prefix_len: 1
  service:
    something: 1
api2:
  exponential_backoff:
    initial_delay: 60s
  service:
    else: 2
`

type api1 struct {
	Something int `yaml:"something"`
}

type api2 struct {
	Else int `yaml:"else"`
}

func TestAPICrawlConfig(t *testing.T) {
	var crawls apicrawlcmd.Crawls
	if err := cmdyaml.ParseConfigString(apiCrawlSpec, &crawls); err != nil {
		t.Fatal(err)
	}

	if got, want := len(crawls), 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	var a1 apicrawlcmd.Crawl[api1]
	if present, err := apicrawlcmd.ParseCrawlConfig(crawls, "api1", &a1); !present || err != nil {
		t.Fatalf("not present: %v, or err: %v", present, err)
	}
	if got, want := a1.Service.Something, 1; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := a1.Cache.Checkpoint, "path/to/checkpoint"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	var a2 apicrawlcmd.Crawl[api2]
	if present, err := apicrawlcmd.ParseCrawlConfig(crawls, "api2", &a2); !present || err != nil {
		t.Fatalf("not present: %v, or err: %v", present, err)
	}

	if got, want := a2.RateControl.ExponentialBackoff.InitialDelay, time.Second*60; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := a2.Service.Else, 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
