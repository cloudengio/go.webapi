// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package biorxivcmd

import (
	"encoding/json"
	"os"
	"time"
)

type crawlState struct {
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Cursor   int64     `json:"cursor"`
	Total    int64     `json:"total"`
	filename string
}

func loadState(filename string) (*crawlState, error) {
	cs := &crawlState{
		filename: filename,
	}
	buf, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return cs, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(buf, cs); err != nil {
		return nil, err
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

func (cs *crawlState) save() error {
	buf, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return err
	}
	tmp := cs.filename + ".tmp"
	os.Remove(tmp)
	if err := os.WriteFile(tmp, buf, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, cs.filename)
}
