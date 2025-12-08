// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package benchlingcmd provides support for building command line tools
// that access benchling.com
package benchlingcmd

import (
	"context"
	"fmt"
	"time"

	"cloudeng.io/errors"
	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/logging/ctxlog"
	"cloudeng.io/path"
	"cloudeng.io/sync/errgroup"
	"cloudeng.io/webapi/clients/benchling"
	"cloudeng.io/webapi/clients/benchling/benchlingsdk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"gopkg.in/yaml.v3"
)

type GetFlags struct{}

type CrawlFlags struct{}

type IndexFlags struct{}

// Çommand implements the command line operations available for protocols.io.
type Command struct {
	state apicrawlcmd.State[Service]
}

// NewCommand returns a new Command instance for benchling API related commands.
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error) {
	state, err := apicrawlcmd.NewState[Service](ctx, config, resources)
	return &Command{state: state}, err
}

func (c *Command) Crawl(ctx context.Context, _ CrawlFlags, entities ...string) error {
	opts, err := OptionsForEndpoint(c.state.Config)
	if err != nil {
		return err
	}
	if err := c.state.Config.Cache.PrepareDownloads(ctx, c.state.Store); err != nil {
		return err
	}
	if err := c.state.Config.Cache.PrepareCheckpoint(ctx, c.state.Checkpoint); err != nil {
		return err
	}
	state, err := loadCheckpoint(ctx, c.state.Checkpoint)
	if err != nil {
		return err
	}
	ctxlog.Info(ctx, "benchling: checkpoint", "user date", state.UsersDate, "entry date", state.EntriesDate)

	ch := make(chan any, 100)

	var crawlGroup errgroup.T
	var errs errors.M
	crawlGroup.Go(func() error {
		return c.crawlSaver(ctx, state, c.state.Config.Cache.DownloadPath(), ch)
	})

	var entityGroup errgroup.T
	for _, entity := range entities {
		entity := entity
		entityGroup.Go(func() error {
			err := c.crawlEntity(ctx, state, entity, ch, opts)
			if err != nil {
				ctxlog.Info(ctx, "benchling: completed crawl of", "entity", entity, "err", err)
			} else {
				ctxlog.Debug(ctx, "benchling: completed crawl of", "entity", entity)
			}
			return err
		})
	}
	errs.Append(entityGroup.Wait())
	close(ch)
	errs.Append(crawlGroup.Wait())
	errs.Append(c.state.Checkpoint.Compact(ctx, ""))
	return errs.Err()
}

func (c *Command) crawlSaver(ctx context.Context, state Checkpoint, downloadsPath string, ch <-chan any) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.state.Config.Cache.ShardingPrefixLen))
	var nUsers, nEntries, nFolders, nProjects int
	var written int64
	defer func() {
		ctxlog.Info(ctx, "benchling: total written", "written", written, "users", nUsers, "entries", nEntries, "folders", nFolders, "projects", nProjects)
	}()
	concurrency := c.state.Config.Cache.Concurrency
	for {
		start := time.Now()
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
		saveStart := time.Now()
		var err error
		switch v := entity.(type) {
		case benchling.Users:
			nUsers += len(v.Users)
			err = save(ctx, c.state.Store, downloadsPath, concurrency, sharder, v.Users)
		case benchling.Entries:
			nEntries += len(v.Entries)
			err = save(ctx, c.state.Store, downloadsPath, concurrency, sharder, v.Entries)
			state.EntriesDate = *(v.Entries[len(v.Entries)-1].ModifiedAt)
			state.UsersDate = time.Now().Format(time.RFC3339)
			if err := saveCheckpoint(ctx, c.state.Checkpoint, state); err != nil {
				ctxlog.Error(ctx, "benchling: failed to save checkpoint", "err", err)
				return err
			}
		case benchling.Folders:
			nFolders += len(v.Folders)
			err = save(ctx, c.state.Store, downloadsPath, concurrency, sharder, v.Folders)
		case benchling.Projects:
			nProjects += len(v.Projects)
			err = save(ctx, c.state.Store, downloadsPath, concurrency, sharder, v.Projects)
		}
		total := nUsers + nEntries + nFolders + nProjects
		ctxlog.Info(ctx, "benchling: written", "total", total, "users", nUsers, "entries", nEntries, "folders", nFolders, "projects", nProjects, "crawl", saveStart.Sub(start), "save", time.Since(saveStart))
		if err != nil {
			return err
		}
	}
}

func (c *Command) crawlEntity(ctx context.Context, state Checkpoint, entity string, ch chan<- any, opts []operations.Option) error {
	switch entity {
	case "users":
		params := c.state.Config.Service.ListUsersConfig()
		from := "> " + state.UsersDate
		params.ModifiedAt = &from
		cr := &crawler[benchling.Users, *benchlingsdk.ListUsersParams]{
			serviceURL: c.state.Config.Service.ServiceURL,
			params:     params,
		}
		return cr.run(ctx, ch, opts)
	case "entries":
		params := c.state.Config.Service.ListEntriesConfig()
		from := "> " + state.EntriesDate
		params.ModifiedAt = &from
		cr := &crawler[benchling.Entries, *benchlingsdk.ListEntriesParams]{
			serviceURL: c.state.Config.Service.ServiceURL,
			params:     params,
		}
		return cr.run(ctx, ch, opts)
	case "folders":
		params := c.state.Config.Service.ListFoldersConfig()
		cr := &crawler[benchling.Folders, *benchlingsdk.ListFoldersParams]{
			serviceURL: c.state.Config.Service.ServiceURL,
			params:     params,
		}
		return cr.run(ctx, ch, opts)
	case "projects":
		params := c.state.Config.Service.ListProjectsConfig()
		cr := &crawler[benchling.Projects, *benchlingsdk.ListProjectsParams]{
			serviceURL: c.state.Config.Service.ServiceURL,
			params:     params,
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
	if len(c.serviceURL) == 0 {
		return fmt.Errorf("no service URL configured")
	}
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

func save[ObjectT benchling.Objects](ctx context.Context, fs content.FS, root string, concurrency int, sharder path.Sharder, obj []ObjectT) error {
	store := stores.New(fs, concurrency)
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
	return store.Finish(ctx)
}

// CreateIndexableDocuments constructs the documents to be indexed from the
// various objects crawled from the benchling.com API.
func (c *Command) CreateIndexableDocuments(ctx context.Context, _ IndexFlags) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.state.Config.Cache.ShardingPrefixLen))
	nd := benchling.NewDocumentIndexer(c.state.Store, c.state.Config.Cache.DownloadPath(), sharder, c.state.Config.Cache.Concurrency)
	return nd.Index(ctx)
}
