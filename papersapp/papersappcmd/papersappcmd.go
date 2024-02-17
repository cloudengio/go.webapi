// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersappcmd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"sync"

	"cloudeng.io/cmdutil/cmdyaml"
	"cloudeng.io/file/content"
	"cloudeng.io/file/content/stores"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/path"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"cloudeng.io/webapi/operations/apitokens"
	"cloudeng.io/webapi/papersapp"
	"cloudeng.io/webapi/papersapp/papersappsdk"
)

type CommonFlags struct {
	PapersAppConfig string `subcmd:"papersapp-config,$HOME/.papersapp.yaml,'papersapp.com auth config file'"`
}

type CrawlFlags struct {
	CommonFlags
}

type ScanFlags struct {
	CommonFlags
	GZIPArchive string `subcmd:"gzip-archive,'','write scanned items to a gzip archive'"`
}

// Çommand implements the command line operations available for papersapp.com.
type Command struct {
	Auth
	Config
	cfs operations.FS
}

// NewCommand returns a new Command instance for the specified API crawl
// with API authentication information read from the specified file or
// from the context.
func NewCommand(ctx context.Context, crawls apicrawlcmd.Crawls, cfs operations.FS, name, authFilename string) (*Command, error) {
	c := &Command{cfs: cfs}
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

func (c *Command) Crawl(ctx context.Context, cacheRoot string, _ *CrawlFlags) error {
	opts, err := c.OptionsForEndpoint(c.Auth)
	if err != nil {
		return err
	}

	_, downloadsPath, _ := c.Cache.AbsolutePaths(c.cfs, cacheRoot)
	if err := c.Cache.PrepareDownloads(ctx, c.cfs, downloadsPath); err != nil {
		return err
	}

	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	collections, err := papersapp.ListCollections(ctx, c.Service.ServiceURL, opts...)
	if err != nil {
		return err
	}

	collectionsCache := stores.New(c.cfs, c.Cache.Concurrency)
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
	fmt.Printf("crawled %v collections\n", len(collections))

	for _, col := range collections {
		if !col.Shared {
			continue
		}
		crawler := &crawlCollection{
			Config:     c.Config,
			fs:         c.cfs,
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
	Config
	fs         content.FS
	root       string
	sharder    path.Sharder
	opts       []operations.Option
	collection *papersappsdk.Collection
}

func (cc *crawlCollection) run(ctx context.Context) error {
	written := 0
	store := stores.New(cc.fs, cc.Cache.Concurrency)
	defer func() {
		store.Finish(ctx)
		log.Printf("total written: %v: %v\n", written, cc.collection.Name)
	}()

	join := cc.fs.Join
	var pgOpts papersapp.ItemPaginatorOptions
	pgOpts.EndpointURL = cc.Service.ServiceURL + "/collections/" + cc.collection.ID + "/items"
	if cc.Service.ListItemsPageSize == 0 {
		cc.Service.ListItemsPageSize = 50
	}
	pgOpts.Parameters = url.Values{}
	pgOpts.Parameters.Add("size", fmt.Sprintf("%v", cc.Service.ListItemsPageSize))
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
				log.Printf("written: %v\n", written)
			}
		}
	}
	log.Printf("%v: % 8v (%v): done\n", cc.collection.ID, dl, cc.collection.Name)
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

	return store.ReadV(ctx, prefix, files, func(ctx context.Context, prefix, name string, ctype content.Type, buf []byte, err error) error {
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
			fmt.Printf("collection: %v: %v\n", obj.Value.ID, obj.Value.Name)
		case papersapp.ItemType:
			var obj content.Object[papersapp.Item, operations.Response]
			if err := obj.Decode(buf); err != nil {
				return err
			}
			item := obj.Value
			fmt.Printf("item: %v: %v: %v\n", item.Item.ItemType, item.Item.ID, item.Collection.Name)
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

func (c *Command) ScanDownloaded(ctx context.Context, root string, fv *ScanFlags) error {
	var gzipWriter io.WriteCloser
	if len(fv.GZIPArchive) > 0 {
		f, err := os.Create(fv.GZIPArchive)
		if err != nil {
			return err
		}
		defer f.Close()
		gzipWriter = gzip.NewWriter(f)
		defer gzipWriter.Close()
	}

	_, downloadsPath, _ := c.Cache.AbsolutePaths(c.cfs, root)

	err := filewalk.ContentsOnly(ctx, c.cfs, downloadsPath, func(ctx context.Context, prefix string, contents []filewalk.Entry, err error) error {
		return scanDownloaded(ctx, c.cfs, c.Cache.Concurrency, gzipWriter, prefix, contents, err)
	})
	return err
}
