// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package apicrawlcmd provides support for building command line tools
// that implement API crawls.
package apicrawlcmd

import (
	"path/filepath"

	"cloudeng.io/file/crawl/crawlcmd"
	"gopkg.in/yaml.v3"
)

// Crawl is a generic type that defines common crawl configuration
// options as well as allowing for service specific ones.
type Crawl[T any] struct {
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:",inline"`
	Service     T                         `yaml:"service" cmd:"service specific configuration"`
}

// Crawls represents the configuration of multiple API crawls.
type Crawls map[string]Crawl[yaml.Node]

// ParseCrawlConfig parses an API specific crawl config of the specified name.
func ParseCrawlConfig[T any](cfg Crawl[yaml.Node], service *Crawl[T]) error {
	service.Cache = cfg.Cache
	service.RateControl = cfg.RateControl
	if err := cfg.Service.Decode(&service.Service); err != nil {
		return err
	}
	return nil
}

// CheckpointPaths returns the paths of all checkpoint directories.
func CheckpointPaths(crawls Crawls) []string {
	var paths []string
	for _, crawl := range crawls {
		if crawl.Cache.Checkpoint != "" {
			paths = append(paths, filepath.Clean(crawl.Cache.Checkpoint))
		}
	}
	return paths
}

// CachePaths returns the paths of all cache directories.
func CachePaths(crawls Crawls) []string {
	var paths []string
	for _, crawl := range crawls {
		if crawl.Cache.Prefix != "" {
			paths = append(paths, filepath.Clean(crawl.Cache.Prefix))
		}
	}
	return paths
}
