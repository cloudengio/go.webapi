// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersappcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"sync"

	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/path"
	"cloudeng.io/webapi/apis/papersapp"
	"cloudeng.io/webapi/apis/papersapp/papersappsdk"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"gopkg.in/yaml.v3"
)

type CrawlFlags struct{}

type ScanFlags struct{}

// Çommand implements the command line operations available for papersapp.com.
type Command struct {
	state apicrawlcmd.State[Service]
}

// NewCommand returns a new Command instance for readcube/papersapp
// API related commands.
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error) {
	state, err := apicrawlcmd.NewState[Service](ctx, config, resources)
	return &Command{state: state}, err
}

func (c *Command) Crawl(ctx context.Context, _ *CrawlFlags) error {
	opts, err := OptionsForEndpoint(c.state.Config, c.state.Token)
	if err != nil {
		return err
	}

	downloadsPath, _ := c.state.Config.Cache.Paths()
	if err := c.state.Config.Cache.PrepareDownloads(ctx, c.state.Store); err != nil {
		return err
	}

	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.state.Config.Cache.ShardingPrefixLen))

	collections, err := papersapp.ListCollections(ctx, c.state.Config.Service.ServiceURL, opts...)
	if err != nil {
		return err
	}

	collectionsCache := stores.New(c.state.Store, c.state.Config.Cache.Concurrency)
	for _, col := range collections {
		obj := content.Object[*papersappsdk.Collection, operations.Response]{
			Type:     papersapp.CollectionType,
			Value:    col,
			Response: operations.Response{},
		}
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", col.ID))
		prefix = collectionsCache.FS().Join(downloadsPath, prefix)
		if err := obj.Store(ctx, collectionsCache, prefix, suffix, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
			return err
		}
	}
	if err := collectionsCache.Finish(ctx); err != nil {
		return err
	}
	log.Printf("papersapp: crawled %v collections\n", len(collections))

	for _, col := range collections {
		if !col.Shared {
			continue
		}
		crawler := &crawlCollection{
			config:     c.state.Config,
			fs:         c.state.Store,
			root:       downloadsPath,
			sharder:    sharder,
			collection: col,
			opts:       opts,
		}
		if err := crawler.run(ctx); err != nil {
			return err
		}
	}
	return nil
}

type crawlCollection struct {
	config     apicrawlcmd.Crawl[Service]
	fs         content.FS
	root       string
	sharder    path.Sharder
	opts       []operations.Option
	collection *papersappsdk.Collection
}

func (cc *crawlCollection) run(ctx context.Context) error {
	written := 0
	store := stores.New(cc.fs, cc.config.Cache.Concurrency)
	defer func() {
		store.Finish(ctx) //nolint:errcheck
		log.Printf("papersapp: total written: %v: %v\n", written, cc.collection.Name)
	}()

	join := cc.fs.Join
	var pgOpts papersapp.ItemPaginatorOptions
	pgOpts.EndpointURL = cc.config.Service.ServiceURL + "/collections/" + cc.collection.ID + "/items"
	if cc.config.Service.ListItemsPageSize == 0 {
		cc.config.Service.ListItemsPageSize = 50
	}
	pgOpts.Parameters = url.Values{}
	pgOpts.Parameters.Add("size", fmt.Sprintf("%v", cc.config.Service.ListItemsPageSize))
	pg := papersapp.NewItemPaginator(pgOpts)
	dl := 0
	sc := operations.NewScanner(pg, cc.opts...)
	for sc.Scan(ctx) {
		items := sc.Response()
		var resp operations.Response
		resp.FromHTTPResponse(sc.HTTPResponse())
		dl += len(items.Items)
		for _, item := range items.Items {
			obj := content.Object[papersapp.Item, operations.Response]{
				Type: papersapp.ItemType,
				Value: papersapp.Item{
					Item:       item,
					Collection: cc.collection,
				},
				Response: resp,
			}
			prefix, suffix := cc.sharder.Assign(fmt.Sprintf("%v", item.ID))
			prefix = join(cc.root, prefix)
			if err := obj.Store(ctx, store, prefix, suffix, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
				return err
			}
			written++
			if written%100 == 0 {
				log.Printf("papersapp: written: %v\n", written)
			}
		}
	}
	log.Printf("papersapp: %v: % 8v (%v): done\n", cc.collection.ID, dl, cc.collection.Name)
	if err := sc.Err(); err != nil {
		return err
	}
	return store.Finish(ctx)
}

func scanDownloaded(ctx context.Context, fs content.FS, concurrency int, gzipWriter io.WriteCloser, prefix string, contents []filewalk.Entry, err error) error {
	if err != nil {
		if fs.IsNotExist(err) {
			return nil
		}
		return err
	}
	store := stores.New(fs, concurrency)
	files := make([]string, len(contents))
	for i, file := range contents {
		files[i] = file.Name
	}
	var mu sync.Mutex

	return store.ReadV(ctx, prefix, files, func(_ context.Context, _, _ string, ctype content.Type, buf []byte, err error) error {
		if err != nil {
			return err
		}
		mu.Lock()
		defer mu.Unlock()
		switch ctype {
		case papersapp.CollectionType:
			var obj content.Object[papersappsdk.Collection, operations.Response]
			if err := obj.Decode(buf); err != nil {
				return err
			}
			log.Printf("papersapp: collection: %v: %v\n", obj.Value.ID, obj.Value.Name)
		case papersapp.ItemType:
			var obj content.Object[papersapp.Item, operations.Response]
			if err := obj.Decode(buf); err != nil {
				return err
			}
			item := obj.Value
			log.Printf("papersapp: item: %v: %v: %v\n", item.Item.ItemType, item.Item.ID, item.Collection.Name)
			if gzipWriter != nil {
				buf, err := json.Marshal(item)
				if err != nil {
					return err
				}
				if _, err := gzipWriter.Write(buf); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (c *Command) ScanDownloaded(ctx context.Context, _ *ScanFlags) error {
	downloadsPath, _ := c.state.Config.Cache.Paths()
	err := filewalk.ContentsOnly(ctx, c.state.Store, downloadsPath, func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
		return scanDownloaded(ctx, c.state.Store, c.state.Config.Cache.Concurrency, nil, prefix, contents, err)
	})
	return err
}
