// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package benchlingcmd provides support for building command line tools
// that access benchling.com
package benchlingcmd

import (
	"context"
	"fmt"

	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/errors"
	"cloudeng.io/file/content"
	"cloudeng.io/path"
	"cloudeng.io/sync/errgroup"
	"cloudeng.io/webapi/benchling"
	"cloudeng.io/webapi/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"cloudeng.io/webapi/operations/apitokens"
)

type CommonFlags struct {
	BenchlingConfig string `subcmd:"benchling-config,$HOME/.benchling.yaml,'benchling.io auth config file'"`
}

type GetFlags struct {
	CommonFlags
}

type CrawlFlags struct {
	CommonFlags
}

type IndexFlags struct {
	CommonFlags
}

// Çommand implements the command line operations available for protocols.io.
type Command struct {
	Auth
	Config
	cfs       operations.FS
	cacheRoot string
}

// NewCommand returns a new Command instance for the specified API crawl
// with API authentication information read from the specified file or
// from the context.
func NewCommand(ctx context.Context, crawls apicrawlcmd.Crawls, cfs operations.FS, cacheRoot, name, authFilename string) (*Command, error) {
	c := &Command{cfs: cfs, cacheRoot: cacheRoot}
	ok, err := apicrawlcmd.ParseCrawlConfig(crawls, name, (*apicrawlcmd.Crawl[Service])(&c.Config))
	if !ok {
		return nil, fmt.Errorf("no configuration found for %v", name)
	}
	if err != nil {
		return nil, err
	}
	if len(authFilename) > 0 {
		if err := cmdyaml.ParseConfigFile(ctx, authFilename, &c.Auth); err != nil {
			return nil, err
		}
	} else {
		if ok, err := apitokens.ParseTokensYAML(ctx, name, &c.Auth); ok && err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Command) Crawl(ctx context.Context, _ CrawlFlags, entities ...string) error {
	opts, err := c.OptionsForEndpoint(c.Auth)
	if err != nil {
		return err
	}
	_, downloadsPath, _ := c.Cache.AbsolutePaths(c.cfs, c.cacheRoot)
	if err := c.Cache.PrepareDownloads(ctx, c.cfs, downloadsPath); err != nil {
		return err
	}

	errCh := make(chan error)
	ch := make(chan any, 100)

	var errs errors.M
	go func() {
		errCh <- c.crawlSaver(ctx, downloadsPath, ch)
	}()

	var g errgroup.T
	for _, entity := range entities {
		entity := entity
		g.Go(func() error {
			return c.crawlEntity(ctx, entity, ch, opts)
		})
	}
	errs.Append(g.Wait())
	close(ch)
	errs.Append(<-errCh)
	return errs.Err()
}

func (c *Command) crawlSaver(ctx context.Context, downloadsPath string, ch <-chan any) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))
	store := content.NewStore(c.cfs)
	for {
		var entity any
		var ok bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entity, ok = <-ch:
			if !ok {
				return nil
			}
		}
		var err error
		switch v := entity.(type) {
		case benchling.Users:
			fmt.Printf("users: %v: %v\n", v.NextToken, len(v.Users))
			err = save(ctx, store, downloadsPath, sharder, v.Users)
		case benchling.Entries:
			fmt.Printf("entries: %v: %v\n", v.NextToken, len(v.Entries))
			err = save(ctx, store, downloadsPath, sharder, v.Entries)
		case benchling.Folders:
			fmt.Printf("folders: %v: %v\n", v.NextToken, len(v.Folders))
			err = save(ctx, store, downloadsPath, sharder, v.Folders)
		case benchling.Projects:
			fmt.Printf("projects: %v: %v\n", v.NextToken, len(v.Projects))
			err = save(ctx, store, downloadsPath, sharder, v.Projects)
		}
		if err != nil {
			return err
		}
	}
}

func (c *Command) crawlEntity(ctx context.Context, entity string, ch chan<- any, opts []operations.Option) error {
	switch entity {
	case "users":
		cr := &crawler[benchling.Users, *benchlingsdk.ListUsersParams]{
			serviceURL: c.Service.ServiceURL,
			params:     c.ListUsersConfig(ctx),
		}
		return cr.run(ctx, ch, opts)
	case "entries":
		cr := &crawler[benchling.Entries, *benchlingsdk.ListEntriesParams]{
			serviceURL: c.Service.ServiceURL,
			params:     c.ListEntriesConfig(ctx),
		}
		return cr.run(ctx, ch, opts)
	case "folders":
		cr := &crawler[benchling.Folders, *benchlingsdk.ListFoldersParams]{
			serviceURL: c.Service.ServiceURL,
			params:     c.ListFoldersConfig(ctx),
		}
		return cr.run(ctx, ch, opts)
	case "projects":
		cr := &crawler[benchling.Projects, *benchlingsdk.ListProjectsParams]{
			serviceURL: c.Service.ServiceURL,
			params:     c.ListProjectsConfig(ctx),
		}
		return cr.run(ctx, ch, opts)
	default:
		return fmt.Errorf("unknown entity %v", entity)
	}
}

type crawler[ScannerT benchling.Scanners, ParamsT benchling.Params] struct {
	serviceURL string
	params     ParamsT
}

func (c *crawler[ScannerT, ParamsT]) run(ctx context.Context, ch chan<- any, opts []operations.Option) error {
	sc := benchling.NewScanner[ScannerT](ctx, c.serviceURL, c.params, opts...)
	for sc.Scan(ctx) {
		u := sc.Response()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- u:
		}
	}
	return sc.Err()
}

func save[ObjectT benchling.Objects](ctx context.Context, store *content.Store, root string, sharder path.Sharder, obj []ObjectT) error {
	for _, o := range obj {
		id := benchling.ObjectID(o)
		obj := content.Object[ObjectT, *operations.Response]{
			Type:     benchling.ContentType(o),
			Value:    o,
			Response: &operations.Response{},
		}
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", id))
		prefix = store.FS().Join(root, prefix)
		if err := obj.Store(ctx, store, prefix, suffix, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
			fmt.Printf("failed to write object id %v: %v %v: %v\n", id, prefix, suffix, err)
		}
	}
	return nil
}

// CreateIndexableDocuments constructs the documents to be indexed from the
// various objects crawled from the benchling.com API.
func (c *Command) CreateIndexableDocuments(ctx context.Context, _ IndexFlags) error {
	_, downloads, _ := c.Cache.AbsolutePaths(c.cfs, c.cacheRoot)
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))
	nd := benchling.NewDocumentIndexer(c.cfs, downloads, sharder)
	return nd.Index(ctx)
}
