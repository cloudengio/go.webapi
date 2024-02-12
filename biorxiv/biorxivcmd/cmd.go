// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package biorxivcmd provides support for building command line tools
// that access api.biorxiv.com
package biorxivcmd

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"cloudeng.io/errors"
	"cloudeng.io/file/checkpoint"
	"cloudeng.io/file/content"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/path"
	"cloudeng.io/webapi/biorxiv"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
)

type CommonFlags struct {
}

type GetFlags struct {
	CommonFlags
}

type CrawlFlags struct {
	CommonFlags
	Restart bool `subcmd:"restart,false,'restart the crawl, ignoring the saved checkpoint'"`
}

type ScanFlags struct {
	CommonFlags
	Template string `subcmd:"template,'{{.PreprintDOI}} {{.PreprintTitle}}',template to use for printing fields in the downloaded Preprint objects"`
}

type LookupFlags struct {
	CommonFlags
	Template string `subcmd:"template,'{{.}}',template to use for printing fields in the downloaded Preprint objects"`
}

type IndexFlags struct {
	CommonFlags
}

// Çommand implements the command line operations available for api.biorxiv.org.
type Command struct {
	Auth
	Config
	cfs       operations.FS
	ckpt      checkpoint.Operation
	cacheRoot string
}

// NewCommand returns a new Command instance for the specified API crawl.
func NewCommand(_ context.Context, crawls apicrawlcmd.Crawls, cfs operations.FS, cacheRoot string, chkp checkpoint.Operation, name string) (*Command, error) {
	c := &Command{cfs: cfs, ckpt: chkp, cacheRoot: cacheRoot}
	ok, err := apicrawlcmd.ParseCrawlConfig(crawls, name, (*apicrawlcmd.Crawl[Service])(&c.Config))
	if !ok {
		return nil, fmt.Errorf("no crawl configuration found for %v", name)
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Crawl implements the crawl command. The crawl is incremental and utilizes
// an internal state file to track progress and restart from that point in
// a subsequent crawl. This makes it possible to have a start date that
// predates the creation of biorxiv and an end date of 'now' with each
// incremental crawl picking up where the previous one left off assuming
// that biorxiv doesn't add new preprints with dates that predate the
// current one.
func (c *Command) Crawl(ctx context.Context, flags CrawlFlags) error {
	opts, err := c.OptionsForEndpoint(c.Auth)
	if err != nil {
		return err
	}
	_, downloadsPath, chkptPath := c.Cache.AbsolutePaths(c.cfs, c.cacheRoot)
	if err := c.Cache.PrepareDownloads(ctx, c.cfs, downloadsPath); err != nil {
		return err
	}
	if err := c.Cache.PrepareCheckpoint(ctx, c.ckpt, chkptPath); err != nil {
		return err
	}

	crawlState, err := loadState(ctx, c.ckpt)
	if err != nil {
		return err
	}
	crawlState.sync(flags.Restart, time.Time(c.Config.Service.StartDate), time.Time(c.Config.Service.EndDate))

	log.Printf("starting crawl from %v to %v, cursor: %v\n", crawlState.From, crawlState.To, crawlState.Cursor)

	errCh := make(chan error)
	ch := make(chan biorxiv.Response, 10)

	store := content.NewStore(c.cfs)
	go func() {
		errCh <- c.crawlSaver(ctx, ch, crawlState, store, downloadsPath)
	}()

	sc := biorxiv.NewScanner(c.Config.Service.ServiceURL, crawlState.From, crawlState.To, crawlState.Cursor, opts...)
	for sc.Scan(ctx) {
		ch <- sc.Response()
	}
	close(ch)
	var errs errors.M
	errs.Append(sc.Err())
	err = <-errCh
	if err != nil && strings.Contains(err.Error(), "no articles found for") {
		err = nil
	}
	errs.Append(err)
	return errs.Err()
}

func (c *Command) crawlSaver(ctx context.Context, ch <-chan biorxiv.Response, cs crawlState, store *content.Store, root string) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	join := store.FS().Join
	defer func() {
		_, written := store.Stats()
		log.Printf("total written: %v\n", written)
	}()
	for {
		var resp biorxiv.Response
		var ok bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case resp, ok = <-ch:
			if !ok {
				return nil
			}
		}
		if len(resp.Messages) == 0 {
			continue
		}
		msg := resp.Messages[0]
		if msg.Status != "ok" {
			return fmt.Errorf("unexpected status: %v", msg.Status)
		}
		for _, preprint := range resp.Collection {
			obj := content.Object[biorxiv.PreprintDetail, struct{}]{
				Type:     biorxiv.PreprintType,
				Value:    preprint,
				Response: struct{}{},
			}
			preprint.PreprintDOI = strings.TrimSpace(preprint.PreprintDOI)
			preprint.PublishedDOI = strings.TrimSpace(preprint.PublishedDOI)
			prefix, suffix := sharder.Assign(fmt.Sprintf("%v", preprint.PreprintDOI))
			prefix = join(root, prefix)
			if err := obj.Store(ctx, store, prefix, suffix, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
				return err
			}
			if _, written := store.Stats(); written%100 == 0 {
				log.Printf("written: %v\n", written)
			}
		}
		var cursor int64
		switch v := msg.Cursor.(type) {
		case string:
			cursor, _ = strconv.ParseInt(v, 10, 64)
		case int:
			cursor = int64(v)
		}
		cs.update(cursor+msg.Count, msg.Total)
		if err := cs.save(ctx, c.ckpt); err != nil {
			return err
		}
	}
}

func scanDownloaded(ctx context.Context, store *content.Store, tpl *template.Template, prefix string, contents []filewalk.Entry, err error) error {
	if err != nil {
		return err
	}
	for _, entry := range contents {
		var obj content.Object[biorxiv.PreprintDetail, struct{}]
		_, err := obj.Load(ctx, store, prefix, entry.Name)
		if err != nil {
			return fmt.Errorf("%v %v: %v", prefix, entry.Name, err)
		}
		if err := tpl.Execute(os.Stdout, obj.Value); err != nil {
			return err
		}
		fmt.Printf("\n")
	}
	return nil
}

// ScanDownloaded scans downloaded preprints printing out fields using
// the specified template.
func (c *Command) ScanDownloaded(ctx context.Context, fv *ScanFlags) error {
	tpl, err := template.New("biorxiv").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}
	_, downloads, _ := c.Cache.AbsolutePaths(c.cfs, c.cacheRoot)
	_, downloadsRel, _ := c.Cache.RelativePaths(c.cacheRoot)
	_ = downloadsRel
	store := content.NewStore(c.cfs)
	return filewalk.ContentsOnly(ctx, c.cfs, downloads,
		func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
			return scanDownloaded(ctx, store, tpl, prefix, contents, err)
		})
}

// LookupDownloaded looks up the specified preprints via their 'PreprintDOI'
// printing out fields using the specified template.
func (c *Command) LookupDownloaded(ctx context.Context, fv *LookupFlags, dois ...string) error {
	tpl, err := template.New("biorxiv").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}

	_, downloads, _ := c.Cache.AbsolutePaths(c.cfs, c.cacheRoot)
	store := content.NewStore(c.cfs)

	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	for _, doi := range dois {
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", doi))
		prefix = store.FS().Join(downloads, prefix)
		var obj content.Object[biorxiv.PreprintDetail, struct{}]
		_, err := obj.Load(ctx, store, prefix, suffix)
		if err != nil {
			return err
		}
		if err := tpl.Execute(os.Stdout, obj.Value); err != nil {
			return err
		}
		fmt.Printf("\n")
	}
	return nil
}
