// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package benchlingcmd

import (
	"context"
	"encoding/json"
	"time"

	"cloudeng.io/file/checkpoint"
)

type Checkpoint struct {
	// Dates in rfc.3339 format.
	UsersDate   string `json:"users_date"`
	EntriesDate string `json:"entries_date"`
}

var dayZero = time.Time{}.Format(time.RFC3339)

func loadCheckpoint(ctx context.Context, op checkpoint.Operation) (Checkpoint, error) {
	var cp Checkpoint
	buf, err := op.Latest(ctx)
	if err != nil {
		return cp, err
	}
	if len(buf) > 0 {
		if err := json.Unmarshal(buf, &cp); err != nil {
			return cp, err
		}
	}
	if len(cp.UsersDate) == 0 {
		cp.UsersDate = dayZero
	}
	if len(cp.EntriesDate) == 0 {
		cp.EntriesDate = dayZero
	}
	return cp, nil
}

func saveCheckpoint(ctx context.Context, op checkpoint.Operation, cp Checkpoint) error {
	buf, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	_, err = op.Checkpoint(ctx, "", buf)
	return err
}
