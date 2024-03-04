// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"cloudeng.io/cmdutil"
	"cloudeng.io/net/ratecontrol"
	"cloudeng.io/webapi/apis/benchling"
	"cloudeng.io/webapi/apis/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
)

const serviceURL = "https://altoslabstest.benchling.com/api/v2/"

func main() {
	ctx := context.Background()
	cfg := struct {
		Key string `yaml:"api_key"`
	}{}
	cmdutil.ParseYAMLConfigFile(ctx, "/Users/cnicolaou/.benchling.yaml", &cfg)

	pageSize := 1
	params := benchlingsdk.ListUsersParams{
		PageSize: &pageSize,
	}

	rc := ratecontrol.New(
		ratecontrol.WithCustomBackoff(func() ratecontrol.Backoff {
			return benchling.NewBackoff(time.Minute, 10)
		}))
	sc := benchling.NewScanner[benchling.Users](ctx,
		serviceURL,
		&params,
		operations.WithAuth(benchling.APIToken{Token: cfg.Key}),
		operations.WithRateController(rc, 429),
	)

	for sc.Scan(ctx) {
		u := sc.Response()
		fmt.Printf("%v\n", *u.Users[0].Id)
	}

	err := sc.Err()
	if err != nil {
		panic(err)
	}

}
