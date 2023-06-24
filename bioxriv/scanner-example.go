// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"cloudeng.io/webapi/biorxiv"
)

const serviceURL = "https://api.biorxiv.org/pubs/biorxiv"

func main() {
	ctx := context.Background()

	from, _ := time.Parse("2006-01-02", "2000-01-01")
	sc := biorxiv.NewScanner(serviceURL, from, time.Now(), 0)

	for sc.Scan(ctx) {
		u := sc.Response()
		fmt.Printf("%v %v %v\n", u.Messages[0].Cursor, u.Messages[0].Count, u.Messages[0].Total)
	}
	err := sc.Err()
	if err != nil {
		panic(err)
	}
}
