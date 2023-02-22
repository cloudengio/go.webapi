// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio_test

import (
	"testing"

	"cloudeng.io/cmdutil"
	"cloudeng.io/webapi/protocolsio"
)

const protocolsioSpec = `
auth:
  public_token: token
  public_clientid: clientid
  public_secret: clientsecret
endpoints:
  list_protocols_v3: https://www.protocols.io/api/v3/protocols
  get_protocol_v4: https://www.protocols.io/api/v4/protocols
`

func TestConfig(t *testing.T) {
	var cfg protocolsio.Config
	if err := cmdutil.ParseYAMLConfigString(protocolsioSpec, &cfg); err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.Auth.PublicToken, "token"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := cfg.Endpoints.GetProtocolV4, "https://www.protocols.io/api/v4/protocols"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
