// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package biorxivcmd

import (
	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/net/ratecontrol"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

// Service represents biorxiv specific configuration parameters.
type Service struct {
	ServiceURL string           `yaml:"service_url" cmd:"rxiv service URL, eg. https://api.biorxiv.org/pubs/biorxiv for biorxiv"`
	StartDate  cmdyaml.FlexTime `yaml:"start_date" cmd:"start date for crawl, eg. 2020-01-01"`
	EndDate    cmdyaml.FlexTime `yaml:"end_date" cmd:"end date for crawl, eg. 2020-12-01"`
	// Note, that the Cursor value is generally obtained a from a checkpoint file.
}

// OptionsForEndpoint returns the operations.Option's derived from the
// apicrawlcmd configuration.
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error) {
	opts := []operations.Option{}
	rateCfg := cfg.RateControl
	rcopts := []ratecontrol.Option{}
	if rateCfg.Rate.BytesPerTick > 0 {
		rcopts = append(rcopts, ratecontrol.WithBytesPerTick(rateCfg.Rate.Tick, rateCfg.Rate.BytesPerTick))
	}
	if rateCfg.Rate.RequestsPerTick > 0 {
		rcopts = append(rcopts, ratecontrol.WithRequestsPerTick(rateCfg.Rate.Tick, rateCfg.Rate.RequestsPerTick))
	}
	if rateCfg.ExponentialBackoff.InitialDelay > 0 {
		rcopts = append(rcopts,
			ratecontrol.WithExponentialBackoff(rateCfg.ExponentialBackoff.InitialDelay, rateCfg.ExponentialBackoff.Steps))
	}
	rc := ratecontrol.New(rcopts...)
	opts = append(opts, operations.WithRateController(rc, cfg.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
