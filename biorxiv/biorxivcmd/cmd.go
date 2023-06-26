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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cloudeng.io/errors"
	"cloudeng.io/file/content"
	"cloudeng.io/path"
	"cloudeng.io/webapi/biorxiv"
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
	cachePath string
	sharder   path.Sharder
}

// NewCommand returns a new Command instance for the specified API crawl.
func NewCommand(_ context.Context, crawls apicrawlcmd.Crawls, name string) (*Command, error) {
	c := &Command{}
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
func (c *Command) Crawl(ctx context.Context, cacheRoot string, flags CrawlFlags) error {
	opts, err := c.OptionsForEndpoint(c.Auth)
	if err != nil {
		return err
	}
	cachePath, checkpointPath, err := c.Cache.Initialize(cacheRoot)
	if err != nil {
		return err
	}
	c.cachePath = cachePath
	c.sharder = path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	crawlState, err := loadState(filepath.Join(checkpointPath, "crawlstate.json"))
	if err != nil {
		return err
	}
	crawlState.sync(flags.Restart, c.Config.Service.StartDate, c.Config.Service.EndDate)

	fmt.Printf("starting crawl from %v to %v, cursor: %v\n", crawlState.From, crawlState.To, crawlState.Cursor)

	errCh := make(chan error)
	ch := make(chan biorxiv.Response, 10)

	go func() {
		errCh <- c.crawlSaver(ctx, ch, crawlState)
	}()

	sc := biorxiv.NewScanner(c.Config.Service.ServiceURL, crawlState.From, crawlState.To, crawlState.Cursor, opts...)
	for sc.Scan(ctx) {
		ch <- sc.Response()
	}
	close(ch)
	var errs errors.M
	errs.Append(sc.Err())
	errs.Append(<-errCh)
	return errs.Err()
}

func (c *Command) crawlSaver(ctx context.Context, ch <-chan biorxiv.Response, cs *crawlState) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

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
			fmt.Printf("writing: %v as %v\n", preprint.PreprintDOI, filepath.Join(prefix, suffix))
			path := filepath.Join(c.cachePath, prefix, suffix)
			if err := WriteDownload(path, obj); err != nil {
				return err
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
		if err := cs.save(); err != nil {
			return err
		}
	}
}

// ScanDownloaded scans downloaded preprints printing out fields using
// the specified template.
func (c *Command) ScanDownloaded(ctx context.Context, root string, fv *ScanFlags) error {
	tpl, err := template.New("biorxiv").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}
	cp := filepath.Join(root, c.Cache.Prefix)
	err = filepath.Walk(cp, func(path string, info os.FileInfo, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return err
		}
		if path == c.Cache.Checkpoint {
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		obj, err := ReadDownload(path)
		if err != nil {
			return err
		}
		if err := tpl.Execute(os.Stdout, obj.Value); err != nil {
			return err
		}
		fmt.Printf("\n")
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return nil
	})
	return err
}

// LookupDownloaded looks up the specified preprints via their 'PreprintDOI'
// printing out fields using the specified template.
func (c *Command) LookupDownloaded(_ context.Context, root string, fv *LookupFlags, dois ...string) error {
	tpl, err := template.New("biorxiv").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}
	cp := filepath.Join(root, c.Cache.Prefix)
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	for _, doi := range dois {
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", doi))
		path := filepath.Join(cp, prefix, suffix)
		obj, err := ReadDownload(path)
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
