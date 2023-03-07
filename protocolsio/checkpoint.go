// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsio

// Checkpoint represents the most recently completed page of the paginated
// crawl.
type Checkpoint struct {
	CompletedPage int64 `json:"completed_page"`
	CurrentPage   int64 `json:"current_page"`
	TotalPages    int64 `json:"total_pages"`
}
