// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchling

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Backoff implements a backoff strategy that first looks for the
// specific backoff period specified in the x-rate-limit-reset header
// in benchling.com's http respone when a rate limit is reached. If no
// such header is found then exponential backoff is used.
type Backoff struct {
	steps     int
	retries   int
	nextDelay time.Duration
	rnd       *rand.Rand
}

func NewBackoff(initial time.Duration, steps int) *Backoff {
	return &Backoff{
		nextDelay: initial,
		steps:     steps,
		rnd:       rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (bb *Backoff) Retries() int {
	return bb.retries
}

// Wait implements Backoff.
func (bb *Backoff) Wait(ctx context.Context, resp *http.Response) (bool, error) {
	if bb.retries >= bb.steps {
		return true, nil
	}
	remaining := resp.Header.Get("x-rate-limit-reset")
	secs, _ := strconv.ParseInt(remaining, 10, 64)
	delay := bb.nextDelay
	if secs > 0 {
		delay = time.Second * time.Duration(secs)
		// Add some jitter.
		delay += time.Second * time.Duration(bb.rnd.Intn(bb.retries+1))
		log.Printf("waiting %v for the current rate limit period to end", delay)
	}
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	case <-time.After(delay):
	}
	bb.retries++
	if secs == 0 {
		bb.nextDelay *= 2
	}
	return false, nil
}
