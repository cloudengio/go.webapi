// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package biorxivcmd

import (
	"context"
	"encoding/json"
	"time"

	"cloudeng.io/file/checkpoint"
)

type crawlState struct {
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
	Cursor int64     `json:"cursor"`
	Total  int64     `json:"total"`
}

func loadState(ctx context.Context, op checkpoint.Operation) (crawlState, error) {
	buf, err := op.Latest(ctx)
	if err != nil || buf == nil {
		return crawlState{}, err
	}
	var cs crawlState
	if err := json.Unmarshal(buf, &cs); err != nil {
		return crawlState{}, err
	}
	return cs, nil
}

func (cs *crawlState) clear() {
	cs.From = time.Time{}
	cs.Cursor = 0
	cs.Total = 0
}

func (cs *crawlState) sync(restart bool, from, to time.Time) {
	if restart {
		cs.clear()
		return
	}
	cs.To = to
	if cs.To.IsZero() {
		cs.To = time.Now()
	}
	if cs.From.Equal(from) {
		return
	}
	cs.From = from
	cs.Cursor = 0
	cs.Total = 0
}

func (cs *crawlState) update(counter, total int64) {
	cs.Cursor = counter
	cs.Total = total
}

func (cs *crawlState) save(ctx context.Context, op checkpoint.Operation) error {
	buf, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return err
	}
	_, err = op.Checkpoint(ctx, "", buf)
	return err
}
