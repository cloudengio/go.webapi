// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchling

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"cloudeng.io/logging/ctxlog"
)

// Backoff implements a backoff strategy that first looks for the
// specific backoff period specified in the x-rate-limit-reset header
// in benchling.com's http response when a rate limit is reached. If no
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
func (bb *Backoff) Wait(ctx context.Context, r any) (bool, error) {
	resp, ok := r.(*http.Response)
	if !ok {
		ctxlog.Error(ctx, "benchling.Backoff.Wait: expected *http.Response, got %T", r)
		return true, fmt.Errorf("expected *http.Response, got %T", r)
	}
	if bb.retries >= bb.steps {
		return true, nil
	}
	delay := bb.nextDelay
	secs := int64(0)
	if resp == nil {
		ctxlog.Info(ctx, "benchling.Backoff.Wait: response is nil")
	}
	if resp != nil && resp.Header == nil {
		ctxlog.Info(ctx, "benchling.Backoff.Wait: response header is nil", "response", resp)
	}
	if resp != nil && resp.Header != nil {
		remaining := resp.Header.Get("x-rate-limit-reset")
		secs, _ = strconv.ParseInt(remaining, 10, 64)
		if secs > 0 {
			delay = time.Second * time.Duration(secs)
			// Add some jitter.
			delay += time.Second * time.Duration(bb.rnd.Intn(bb.retries+1))
			ctxlog.Info(ctx, "benchling.Backoff.Wait: waiting", "delay", delay)
		}
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
