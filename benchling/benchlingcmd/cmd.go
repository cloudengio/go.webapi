// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package benchlingcmd provides support for building command line tools
// that access benchling.com
package benchlingcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"cloudeng.io/cmdutil"
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
	cachePath string
	sharder   path.Sharder
}

// NewCommand returns a new Command instance for the specified API crawl
// with API authentication information read from the specified file or
// from the context.
func NewCommand(ctx context.Context, crawls apicrawlcmd.Crawls, name, authFilename string) (*Command, error) {
	c := &Command{}
	ok, err := apicrawlcmd.ParseCrawlConfig(crawls, name, (*apicrawlcmd.Crawl[Service])(&c.Config))
	if !ok {
		return nil, fmt.Errorf("no configuration found for %v", name)
	}
	if err != nil {
		return nil, err
	}
	if len(authFilename) > 0 {
		if err := cmdutil.ParseYAMLConfigFile(ctx, authFilename, &c.Auth); err != nil {
			return nil, err
		}
	} else {
		if ok, err := apitokens.ParseTokensYAML(ctx, name, &c.Auth); ok && err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Command) Crawl(ctx context.Context, cacheRoot string, flags CrawlFlags, entities ...string) error {
	opts, err := c.OptionsForEndpoint(c.Auth)
	if err != nil {
		return err
	}
	cachePath, _, err := c.Cache.Initialize(cacheRoot)
	if err != nil {
		return err
	}
	c.cachePath = cachePath
	c.sharder = path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	errCh := make(chan error)
	ch := make(chan any, 100)

	var errs errors.M
	go func() {
		errCh <- c.crawlSaver(ctx, ch)
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

func (c *Command) crawlSaver(ctx context.Context, ch <-chan any) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

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
			err = save(ctx, c.cachePath, sharder, v.Users)
		case benchling.Entries:
			fmt.Printf("entries: %v: %v\n", v.NextToken, len(v.Entries))
			err = save(ctx, c.cachePath, sharder, v.Entries)
		case benchling.Folders:
			fmt.Printf("folders: %v: %v\n", v.NextToken, len(v.Folders))
			err = save(ctx, c.cachePath, sharder, v.Folders)
		case benchling.Projects:
			fmt.Printf("projects: %v: %v\n", v.NextToken, len(v.Projects))
			err = save(ctx, c.cachePath, sharder, v.Projects)
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

func save[ObjectT benchling.Objects](ctx context.Context, cachePath string, sharder path.Sharder, obj []ObjectT) error {
	for _, o := range obj {
		id := benchling.ObjectID(o)
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", id))
		path := filepath.Join(cachePath, prefix, suffix)
		obj := content.Object[ObjectT, *operations.Response]{
			Type:     benchling.ContentType(o),
			Value:    o,
			Response: &operations.Response{},
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := obj.WriteObjectFile(path, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
			fmt.Printf("failed to write user: %v as %v: %v\n", id, path, err)
		}
		fmt.Printf("%v\n", path)
	}
	return nil
}

func (c *Command) Index(ctx context.Context, root string, flags IndexFlags, entities ...string) error {
	cp := filepath.Join(root, c.Cache.Prefix)
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))
	nd := benchling.NewDocumentIndexer(cp, sharder)
	return nd.Index(ctx)
}
