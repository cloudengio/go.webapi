// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsiocmd provides support for building command line tools
// that access protocols.io.
package protocolsiocmd

import (
	"context"

	"cloudeng.io/cmdutil/flags"
)

func ListFilters() []string {
	return []string{"public", "user_public", "user_private", "shared_with_user"}
}

func ListOrderField() []string {
	return []string{"activity", "relevance", "date", "name", "id"}
}

func ListOrderDirection() []string {
	return []string{"asc", "desc"}
}

type CommonFlags struct {
	Config string `subcmd:"config,$HOME/.protocolsio.yaml,'protocols.io config file'"`
}

type ListFlags struct {
	Pages    flags.IntRangeSpec `subcmd:"pages,1,page range to return"`
	PageSize int                `subcmd:"size,20,number of items in each page"`
	Total    int                `subcmd:"total,,total number of items to return"`
	Filter   string             `subcmd:"filter,public,'one of: 1. public - list of all public protocols;	2. user_public - list of public protocols that was publiches by concrete user; 3. user_private - list of private protocols that was created by concrete user; 4. shared_with_user - list of public protocols that was shared with concrete user.'"`
}

type GetFlags struct {
	CommonFlags
}

type QueryFlags struct {
	CommonFlags
	Query string `subcmd:"query,,'string may contain any characters, numbers and special symbols. System will seach around protocol name, description, authors. If the search keywords are enclosed into double quotes, then result contains only the exact match of the combined term'"`
	Order string `subcmd:"order,activity,'one of: 1. activity - index of protocol popularity; relevance - returns most relevant to the search key results at the top; date - date of publication; name - protocol name; id - id of protocol.'"`
	Sort  string `subcmd:"sort,asc,one of asc or desc"`
}

type CrawlFlags struct {
	CommonFlags
	ListFlags
}

type Command struct {
	Config
}

func (c *Command) List(ctx context.Context, fv *ListFlags, args []string) error {
	return nil
}

func (c *Command) Crawl(ctx context.Context, fv *CrawlFlags, args []string) error {
	return nil
}

func (c *Command) Get(ctx context.Context, fv *GetFlags, args []string) error {
	return nil
}
