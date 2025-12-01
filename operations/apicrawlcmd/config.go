// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package apicrawlcmd provides support for building command line tools
// that implement API crawls.
package apicrawlcmd

import (
	"context"

	"cloudeng.io/file/checkpoint"
	"cloudeng.io/file/crawl/crawlcmd"
	"cloudeng.io/webapi/operations"
	"gopkg.in/yaml.v3"
)

// Crawl is a generic type that defines common crawl configuration
// options as well as allowing for service specific ones. The type
// of the service specific configuration is generally determined by
// the API being crawled.
type Crawl[T any] struct {
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:"cache"`
	KeyID       string                    `yaml:"key_id" cmd:"identifier of the API key to use for this crawl"`
	Service     T                         `yaml:"service_config" cmd:"service specific configuration"`
}

// Crawls represents the configuration of multiple API crawls.
type Crawls map[string]Crawl[yaml.Node]

// ParseCrawlConfig parses an API specific crawl config, it's parametized
// by the types of the service specific and crawl cache specific data types.
func ParseCrawlConfig[T any](cfg Crawl[yaml.Node], service *Crawl[T]) error {
	service.RateControl = cfg.RateControl
	service.Cache = cfg.Cache
	if err := cfg.Service.Decode(&service.Service); err != nil {
		return err
	}
	return nil
}

// Resources represents the resources typically required to perform an API crawl.
type Resources struct {
	NewOperationsFS func(ctx context.Context, cfg crawlcmd.CrawlCacheConfig) (operations.FS, error)

	NewCheckpointOp func(ctx context.Context, cfg crawlcmd.CrawlCacheConfig) (checkpoint.Operation, error)
}

func (r Resources) CreateResources(ctx context.Context, cfg crawlcmd.CrawlCacheConfig) (store operations.FS, chkpt checkpoint.Operation, err error) {
	if len(cfg.Downloads) > 0 && r.NewOperationsFS != nil {
		store, err = r.NewOperationsFS(ctx, cfg)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(cfg.Checkpoint) > 0 && r.NewCheckpointOp != nil {
		chkpt, err = r.NewCheckpointOp(ctx, cfg)
		if err != nil {
			return nil, nil, err
		}
	}
	return
}

type State[T any] struct {
	Config     Crawl[T]
	Store      operations.FS
	Checkpoint checkpoint.Operation
}

func NewState[T any](ctx context.Context, config Crawl[yaml.Node], resources Resources) (State[T], error) {
	s := State[T]{}
	err := ParseCrawlConfig(config, &s.Config)
	if err != nil {
		return State[T]{}, err
	}
	s.Store, s.Checkpoint, err = resources.CreateResources(ctx, s.Config.Cache)
	if err != nil {
		return State[T]{}, err
	}
	return s, err
}
