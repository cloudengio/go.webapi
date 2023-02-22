// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsio provides support for working with the
// protocols.io API. It currently provides the abilty to
// crawl public protocols.
package protocolsio

import (
	"fmt"
	"strings"

	"cloudeng.io/file/crawl/crawlcmd"
)

// Config represents the configuration information required to
// access and crawl the protocols.io API.
type Config struct {
	Auth struct {
		PublicToken  string `yaml:"public_token"`
		ClientID     string `yaml:"public_clientid"`
		ClientSecret string `yaml:"public_secret"`
	}
	Endpoints struct {
		ListProtocolsV3 string `yaml:"list_protocols_v3"`
		GetProtocolV4   string `yaml:"get_protocol_v4"`
	}
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:"cache"`
}

func (c Config) String() string {
	var out strings.Builder
	out.WriteString("protocolsio:\n")
	if len(c.Auth.PublicToken) > 0 {
		fmt.Fprintf(&out, " public token: **redacted**\n")
	}
	if len(c.Auth.ClientID) > 0 {
		fmt.Fprintf(&out, "    client id: %s\n", c.Auth.ClientID)
	}
	if len(c.Auth.ClientSecret) > 0 {
		fmt.Fprintf(&out, "client secret: **redacted**\n")
	}
	fmt.Fprintf(&out, "endpoint: list v3: %s", c.Endpoints.ListProtocolsV3)
	fmt.Fprintf(&out, "endpoint: get v4: %s", c.Endpoints.GetProtocolV4)
	return out.String()
}
