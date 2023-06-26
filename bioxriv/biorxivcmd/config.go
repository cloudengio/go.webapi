// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package biorxivcmd

import (
	"time"

	"cloudeng.io/net/ratecontrol"
	"cloudeng.io/webapi/benchling"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

type Auth struct {
	// Currently api.biorxiv.org does not require an api key.
}

type Service struct {
	ServiceURL string    `yaml:"service_url"`
	StartDate  time.Time `yaml:"start_date"`
	EndDate    time.Time `yaml:"end_date"`
}

type Config apicrawlcmd.Crawl[Service]

func (c Config) OptionsForEndpoint(auth Auth) ([]operations.Option, error) {
	opts := []operations.Option{}
	rateCfg := c.RateControl
	rcopts := []ratecontrol.Option{}
	if rateCfg.Rate.BytesPerTick > 0 {
		rcopts = append(rcopts, ratecontrol.WithBytesPerTick(rateCfg.Rate.Tick, rateCfg.Rate.BytesPerTick))
	}
	if rateCfg.Rate.RequestsPerTick > 0 {
		rcopts = append(rcopts, ratecontrol.WithRequestsPerTick(rateCfg.Rate.Tick, rateCfg.Rate.RequestsPerTick))
	}
	if rateCfg.ExponentialBackoff.InitialDelay > 0 {
		rcopts = append(rcopts,
			ratecontrol.WithCustomBackoff(
				func() ratecontrol.Backoff {
					return benchling.NewBackoff(rateCfg.ExponentialBackoff.InitialDelay, rateCfg.ExponentialBackoff.Steps)
				}))
	}
	rc := ratecontrol.New(rcopts...)
	opts = append(opts, operations.WithRateController(rc, c.RateControl.ExponentialBackoff.StatusCodes...))
	return opts, nil
}
