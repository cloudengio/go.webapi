// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsiosdk provides a minimal SDK for the protocols.io API.
// See https://apidoc.protocols.io for details.
package protocolsiosdk

import (
	"encoding/json"
)

const (
	ListProtocolsV3Endpoint = "https://www.protocols.io/api/v3/protocols"
	GetProtocolV4Endpoint   = "https://www.protocols.io/api/v4/protocols"
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
	Protocol   Protocol `json:"payload"`
	StatusCode int      `json:"status_code"`
}

type Pagination struct {
	CurrentPage  int64       `json:"current_page"`
	TotalPages   int64       `json:"total_pages"`
	TotalResults int64       `json:"total_results"`
	NextPage     string      `json:"next_page"`
	PrevPage     interface{} `json:"prev_page"`
	PageSize     int64       `json:"page_size"`
	First        int64       `json:"first"`
	Last         int64       `json:"last"`
	ChangedOn    interface{} `json:"changed_on"`
}
