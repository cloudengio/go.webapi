// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersappcmd

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"cloudeng.io/cmdutil"
	"cloudeng.io/file/content"
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

func (c *Command) Crawl(ctx context.Context, cacheRoot string, fv *CrawlFlags) error {
	cachePath, _, err := c.Cache.Initialize(cacheRoot)
	if err != nil {
		return err
	}
	sharder := path.NewSharder(path.WithSHA1PrefixLength(c.Cache.ShardingPrefixLen))

	opts, err := c.OptionsForEndpoint(c.Auth)
	if err != nil {
		return err
	}

	collections, err := papersapp.ListCollections(ctx, c.Service.ServiceURL, opts...)
	if err != nil {
		return err
	}

	for _, col := range collections {
		prefix, suffix := sharder.Assign(fmt.Sprintf("%v", col.ID))
		path := filepath.Join(cachePath, prefix, suffix)
		obj := content.Object[*papersappsdk.Collection, operations.Response]{
			Type:     papersapp.CollectionType,
			Value:    col,
			Response: operations.Response{},
		}
		if err := WriteDownload(path, obj); err != nil {
			return err
		}
	}

	for _, col := range collections {
		if !col.Shared {
			continue
		}
		crawler := &crawlCollection{
			Config:     c.Config,
			cachePath:  cachePath,
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
	cachePath  string
	sharder    path.Sharder
	opts       []operations.Option
	collection *papersappsdk.Collection
}

func (cc *crawlCollection) run(ctx context.Context) error {
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
		log.Printf("%v: % 8v/% 8v (%v)\n", cc.collection.ID, len(items.Items), items.Total, cc.collection.Name)
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
			path := filepath.Join(cc.cachePath, prefix, suffix)
			if err := WriteDownload(path, obj); err != nil {
				return err
			}
		}
	}
	log.Printf("%v: % 8v (%v): done\n", cc.collection.ID, dl, cc.collection.Name)
	return sc.Err()
}

func (c *Command) ScanDownloaded(_ context.Context, root string, fv *ScanFlags) error {

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

	cp := filepath.Join(root, c.Cache.Prefix)
	err := filepath.Walk(cp, func(path string, info os.FileInfo, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return err
		}
		if path == c.Cache.Checkpoint {
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		ctype, buf, err := content.ReadObjectFile(path)
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
	return err
}
