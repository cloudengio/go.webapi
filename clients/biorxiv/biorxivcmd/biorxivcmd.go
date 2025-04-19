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
	"strings"
	"sync"
	"time"

	"cloudeng.io/errors"
	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/logging/ctxlog"
	"cloudeng.io/path"
	"cloudeng.io/webapi/clients/biorxiv"
	util "cloudeng.io/webapi/clients/biorxiv/internal"

	"cloudeng.io/webapi/operations/apicrawlcmd"
	"gopkg.in/yaml.v3"
)

type GetFlags struct{}

type CrawlFlags struct {
	Restart bool `subcmd:"restart,false,'restart the crawl, ignoring the saved checkpoint'"`
}

type ScanFlags struct {
	Template string `subcmd:"template,'{{.PreprintDOI}} {{.PreprintTitle}}',template to use for printing fields in the downloaded Preprint objects"`
}

type LookupFlags struct {
	Template string `subcmd:"template,'{{.}}',template to use for printing fields in the downloaded Preprint objects"`
}

type IndexFlags struct{}

// Çommand implements the command line operations available for api.biorxiv.org.
type Command struct {
	state apicrawlcmd.State[Service]
}

// NewCommand returns a new Command instance for biorxiv API related commands.
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error) {
	state, err := apicrawlcmd.NewState[Service](ctx, config, resources)
	return &Command{state: state}, err
}

// Crawl implements the crawl command. The crawl is incremental and utilizes
// an internal state file to track progress and restart from that point in
// a subsequent crawl. This makes it possible to have a start date that
// predates the creation of biorxiv and an end date of 'now' with each
// incremental crawl picking up where the previous one left off assuming
// that biorxiv doesn't add new preprints with dates that predate the
// current one.
func (c *Command) Crawl(ctx context.Context, flags CrawlFlags) error {
	opts, err := OptionsForEndpoint(c.state.Config)
	if err != nil {
		return err
	}
	downloadPath := c.state.Config.Cache.DownloadPath()
	if err := c.state.Config.Cache.PrepareDownloads(ctx, c.state.Store); err != nil {
		return err
	}
	if err := c.state.Config.Cache.PrepareCheckpoint(ctx, c.state.Checkpoint); err != nil {
		return err
	}

	crawlState, err := loadState(ctx, c.state.Checkpoint)
	if err != nil {
		return err
	}
	crawlState.sync(flags.Restart, time.Time(c.state.Config.Service.StartDate), time.Time(c.state.Config.Service.EndDate))

	ctxlog.Info(ctx, "biorxiv: starting crawl", "from", crawlState.From, "to", crawlState.To, "cursor", crawlState.Cursor)

	errCh := make(chan error)
	ch := make(chan biorxiv.Response, 10)

	go func() {
		errCh <- c.crawlSaver(ctx, ch, crawlState, c.state.Store, downloadPath)
	}()

	sc := biorxiv.NewScanner(c.state.Config.Service.ServiceURL, crawlState.From, crawlState.To, crawlState.Cursor, opts...)
	for sc.Scan(ctx) {
		resp := sc.Response()
		ctxlog.Info(ctx, "biorxiv: crawled", "preprints", len(resp.Collection))
		ch <- resp
	}
	close(ch)
	var errs errors.M
	errs.Append(sc.Err())
	err = <-errCh
	if err != nil && strings.Contains(err.Error(), "no articles found for") {
		err = nil
	}
	errs.Append(err)
	errs.Append(c.state.Checkpoint.Compact(ctx, ""))
	return errs.Err()
}

func (c *Command) crawlSaver(ctx context.Context, ch <-chan biorxiv.Response, cs crawlState, fs content.FS, root string) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.state.Config.Cache.ShardingPrefixLen))

	store := stores.New(fs, c.state.Config.Cache.Concurrency)
	defer store.Finish(ctx) //nolint:errcheck

	written := 0
	join := fs.Join
	defer func() {
		ctxlog.Info(ctx, "biorxiv: total written", "preprints", written)
	}()
	for {
		var resp biorxiv.Response
		var ok bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case resp, ok = <-ch:
			if !ok {
				return store.Finish(ctx)
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
			written++
			if written%100 == 0 {
				ctxlog.Info(ctx, "biorxiv: written", "preprints", written)
			}
		}
		cursor, err := util.AsInt64(msg.Cursor)
		if err != nil {
			return fmt.Errorf("unexpected cursor: %v: %v", msg.Cursor, err)
		}
		total, err := util.AsInt64(msg.Total)
		if err != nil {
			return fmt.Errorf("unexpected total: %v: %v", msg.Total, err)
		}
		cs.update(cursor+msg.Count, total)
		if err := cs.save(ctx, c.state.Checkpoint); err != nil {
			return err
		}
	}
}

func (c *Command) scanDownloaded(ctx context.Context, tpl *template.Template, prefix string, contents []filewalk.Entry, err error) error {
	if err != nil {
		return err
	}
	store := stores.New(c.state.Store, c.state.Config.Cache.Concurrency)
	names := make([]string, len(contents))
	for i, entry := range contents {
		names[i] = entry.Name
	}
	var mu sync.Mutex

	return store.ReadV(ctx, prefix, names, func(_ context.Context, prefix, name string, _ content.Type, data []byte, err error) error {
		if err != nil {
			return err
		}
		mu.Lock()
		defer mu.Unlock()

		var obj content.Object[biorxiv.PreprintDetail, struct{}]
		if err := obj.Decode(data); err != nil {
			return fmt.Errorf("%v %v: %v", prefix, name, err)
		}
		if err := tpl.Execute(os.Stdout, obj.Value); err != nil {
			return err
		}
		fmt.Printf("\n")
		return nil
	})
}

// ScanDownloaded scans downloaded preprints printing out fields using
// the specified template.
func (c *Command) ScanDownloaded(ctx context.Context, fv *ScanFlags) error {
	tpl, err := template.New("biorxiv").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}
	downloadPath := c.state.Config.Cache.DownloadPath()
	return filewalk.ContentsOnly(ctx, c.state.Store, downloadPath,
		func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
			return c.scanDownloaded(ctx, tpl, prefix, contents, err)
		})
}

// LookupDownloaded looks up the specified preprints via their 'PreprintDOI'
// printing out fields using the specified template.
func (c *Command) LookupDownloaded(ctx context.Context, fv *LookupFlags, dois ...string) error {
	tpl, err := template.New("biorxiv").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}

	downloadPath := c.state.Config.Cache.DownloadPath()
	store := stores.New(c.state.Store, 0)

	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.state.Config.Cache.ShardingPrefixLen))

	for _, doi := range dois {
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", doi))
		prefix = store.FS().Join(downloadPath, prefix)
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
