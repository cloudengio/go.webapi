// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package protocolsiocmd provides support for building command line tools
// that access protocols.io.
package protocolsiocmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"text/template"

	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/errors"
	"cloudeng.io/file/checkpoint"
	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/path"
	"cloudeng.io/webapi/clients/protocolsio/protocolsiosdk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"gopkg.in/yaml.v3"
)

type GetFlags struct{}

type CrawlFlags struct {
	Save             bool               `subcmd:"save,true,'save downloaded protocols to disk'"`
	IgnoreCheckpoint bool               `subcmd:"ignore-checkpoint,false,'ignore the checkpoint files'"`
	Pages            flags.IntRangeSpec `subcmd:"pages,,page range to return"`
	PageSize         int                `subcmd:"size,50,number of items in each page"`
	Key              string             `subcmd:"key,,'string may contain any characters, numbers and special symbols. System will search around protocol name, description, authors. If the search keywords are enclosed in double quotes, then result contains only the exact match of the combined term'"`
}

type ScanFlags struct {
	Template string `subcmd:"template,'{{.ID}}',template to use for printing fields in the downloaded Protocol objects"`
}

// Çommand implements the command line operations available for protocols.io.
type Command struct {
	state apicrawlcmd.State[Service]
}

// NewCommand returns a new Command instance for protocols.io API related commands.
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error) {
	state, err := apicrawlcmd.NewState[Service](ctx, config, resources)
	return &Command{state: state}, err
}

func (c *Command) Crawl(ctx context.Context, fv *CrawlFlags) error {
	downloadPath := c.state.Config.Cache.DownloadPath()
	if err := c.state.Config.Cache.PrepareDownloads(ctx, c.state.Store); err != nil {
		return err
	}
	if err := c.state.Config.Cache.PrepareCheckpoint(ctx, c.state.Checkpoint); err != nil {
		return err
	}
	if fv.IgnoreCheckpoint {
		if err := c.state.Checkpoint.Clear(ctx); err != nil {
			return err
		}
	}

	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.state.Config.Cache.ShardingPrefixLen))

	crawler, err := newProtocolCrawler(ctx, c.state.Config, c.state.Store, downloadPath, c.state.Checkpoint, fv, c.state.Token)
	if err != nil {
		return err
	}

	var errs errors.M
	err = operations.RunCrawl(ctx, crawler,
		func(ctx context.Context, objects []content.Object[protocolsiosdk.ProtocolPayload, operations.Response]) error {
			return handleCrawledObject(ctx, fv.Save, sharder, c.state.Store, downloadPath, c.state.Checkpoint, objects)
		})
	errs.Append(err)
	errs.Append(c.state.Checkpoint.Compact(ctx, ""))
	return errs.Err()
}

func handleCrawledObject(ctx context.Context,
	save bool,
	sharder path.Sharder,
	fs content.FS,
	root string,
	chk checkpoint.Operation,
	objs []content.Object[protocolsiosdk.ProtocolPayload, operations.Response]) error {

	store := stores.New(fs, 0)
	for _, obj := range objs {
		if obj.Response.Current != 0 && obj.Response.Total != 0 {
			log.Printf("protocols.io: progress: %v/%v\n", obj.Response.Current, obj.Response.Total)
		}
		if obj.Value.Protocol.ID == 0 {
			// Protocol is up-to-date on disk.
			return nil
		}
		log.Printf("protocols.io: protocol ID: %v\n", obj.Value.Protocol.ID)
		if !save {
			return nil
		}
		// Save the protocol object to disk.
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", obj.Value.Protocol.ID))
		prefix = store.FS().Join(root, prefix)
		if err := obj.Store(ctx, store, prefix, suffix, content.GOBObjectEncoding, content.GOBObjectEncoding); err != nil {
			return err
		}

		if state := obj.Response.Checkpoint; len(state) > 0 {
			name, err := chk.Checkpoint(ctx, "", state)
			if err != nil {
				log.Printf("protocols.io: failed to save checkpoint: %v: %v\n", name, err)
			} else {
				log.Printf("protocols.io: checkpoint: %v\n", name)
			}
		}
	}
	return store.Finish(ctx)
}

func (c *Command) Get(ctx context.Context, _ *GetFlags, args []string) error {
	opts, err := OptionsForEndpoint(c.state.Config, c.state.Token)
	if err != nil {
		return err
	}
	ep := operations.NewEndpoint[protocolsiosdk.ProtocolPayload](opts...)
	for _, id := range args {
		u := fmt.Sprintf("%v/%v", protocolsiosdk.GetProtocolV4Endpoint, id)
		obj, _, _, err := ep.Get(ctx, u)
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", obj)
	}
	return nil
}

func (c *Command) ScanDownloaded(ctx context.Context, fv *ScanFlags) error {
	tpl, err := template.New("protocolsio").Parse(fv.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %q: %v", fv.Template, err)
	}
	downloadPath := c.state.Config.Cache.DownloadPath()
	store := stores.New(c.state.Store, c.state.Config.Cache.Concurrency)
	var mu sync.Mutex
	err = filewalk.ContentsOnly(ctx, c.state.Store, downloadPath, func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
		if err != nil {
			log.Printf("protocols.io: error: %v: %v", prefix, err)
		}
		names := make([]string, len(contents))
		for i, c := range contents {
			names[i] = c.Name
		}
		return store.ReadV(ctx, prefix, names, func(_ context.Context, _, _ string, _ content.Type, buf []byte, err error) error {
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			var obj content.Object[protocolsiosdk.ProtocolPayload, operations.Response]
			if err := obj.Decode(buf); err != nil {
				return err
			}
			if err := tpl.Execute(os.Stdout, obj.Value.Protocol); err != nil {
				return err
			}
			fmt.Printf("\n")
			return nil
		})
	})
	return err
}
