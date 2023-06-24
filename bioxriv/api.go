// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package biorxiv provides support for crawling and indexing biorxiv.org and
// medrxiv.org via api.biorxiv.org.
package biorxiv

import (
	"cloudeng.io/file/content"
)

type PreprintDetail struct {
	PreprintDOI                            string `json:"preprint_doi"`
	PublishedDOI                           string `json:"published_doi"`
	PublishedJournal                       string `json:"published_journal"`
	PreprintPlatform                       string `json:"preprint_platform"`
	PreprintTitle                          string `json:"preprint_title"`
	PreprintAuthors                        string `json:"preprint_authors"`
	PreprintCategory                       string `json:"preprint_category"`
	PreprintDate                           string `json:"preprint_date"`
	PublishedDate                          string `json:"published_date"`
	PreprintAbstract                       string `json:"preprint_abstract"`
	PreprintAuthroCorresponding            string `json:"preprint_author_corresponding"`
	PreprintAuthorCoresspondingInstitution string `json:"preprint_author_corresponding_institution"`
}

type Message struct {
	Status   string `json:"status"`
	Interval string `json:"interval"`
	Cursor   any    `json:"cursor"` // Can be an int or a string!
	Count    int64  `json:"count"`
	Total    int64  `json:"total"`
}

type Response struct {
	Messages   []Message        `json:"messages"`
	Collection []PreprintDetail `json:"collection"`
}

const (
	DocumentType = content.Type("api.biorxiv.org/article")
)
