// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsiocmd_test

import (
	"testing"

	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/webapi/apis/protocolsio/protocolsiocmd"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

const protocolsioSpec = `
protocols.io:
  service:
    order_field: id
  rate_control:
    requests_per_tick: 3
`

func TestConfig(t *testing.T) {
	var crawls apicrawlcmd.Crawls
	if err := cmdyaml.ParseConfigString(protocolsioSpec, &crawls); err != nil {
		t.Fatal(err)
	}
	var cfg apicrawlcmd.Crawl[protocolsiocmd.Service]
	err := apicrawlcmd.ParseCrawlConfig(crawls["protocols.io"], &cfg)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := cfg.Service.OrderField, "id"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := cfg.RateControl.Rate.RequestsPerTick, 3; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
