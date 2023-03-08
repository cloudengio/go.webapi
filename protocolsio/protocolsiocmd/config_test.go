// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsiocmd_test

import (
	"testing"

	"cloudeng.io/cmdutil"
	"cloudeng.io/webapi/protocolsio/protocolsiocmd"
)

const protocolsioSpec = `
service:
  auth:
    public_token: token
    public_clientid: clientid
    public_secret: clientsecret
order_field: id
rate_control:
  requests_per_tick: 3
`

func TestConfig(t *testing.T) {
	var cfg protocolsiocmd.Config
	if err := cmdutil.ParseYAMLConfigString(protocolsioSpec, &cfg); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Service.Auth.PublicToken, "token"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := cfg.Service.OrderField, "id"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := cfg.RateControl.Rate.RequestsPerTick, 3; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
