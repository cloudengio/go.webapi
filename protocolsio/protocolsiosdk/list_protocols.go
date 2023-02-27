// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsiosdk

import (
	"encoding/json"
)

type ListProtocolsV3 struct {
	Extras       json.RawMessage
	Items        []json.RawMessage `json:"items"`
	Pagination   Pagination        `json:"pagination"`
	Total        int64             `json:"total"`
	TotalPages   int64             `json:"total_pages"`
	TotalResults int64             `json:"total_results"`
}

type Payload struct {
	Payload    json.RawMessage `json:"payload"`
	StatusCode int             `json:"status_code"`
}

type Creator struct {
	Name       string
	Username   string
	Affilation string
}

type Protocol struct {
	ID          int64  `json:"id"`
	URI         string `json:"uri"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VersionID   int    `json:"version_id"`
	CreatedOn   int    `json:"created_on"`
	Creator     Creator
}

type ProtocolPayload struct {
	Protocol   Protocol
	StatusCode int `json:"status_code"`
}

/*
func ParsePayload[T any](buf []byte) (T, error) {
	var t T
	var payload Payload
	if err := json.Unmarshal(buf, &payload); err != nil {
		return t, err
	}
	err := json.Unmarshal(payload.Payload, &t)
	return t, err
}
*/
