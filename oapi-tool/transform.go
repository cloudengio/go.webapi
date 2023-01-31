// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloudeng.io/webapi/openapi/transforms"
	"gopkg.in/yaml.v3"
)

type TransformFlags struct {
	OutputFlags
	Config   string `subcmd:"config,transform.yaml,yaml configuration for the transformations to be applied"`
	Describe bool   `subcmd:"describe,,describe all configured transformations"`
	Validate bool   `subcmd:"validate,true,validate the transformed specification"`
}

func transformCmd(ctx context.Context, values any, args []string) error {
	fv := values.(*TransformFlags)
	cfg, err := transforms.LoadConfigFile(fv.Config)
	if err != nil {
		return err
	}

	if fv.Describe {
		return applyTransformations(ctx, cfg, func(ctx context.Context, t transforms.T, node yaml.Node) error {
			out := t.Describe(node)
			fmt.Println(out)
			return nil
		})
	}

	if err := cfg.ConfigureAll(); err != nil {
		return err
	}

	filename := args[0]
	asYAML, err := OutputFormat(filename, fv.ConvertToYAML)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	doc, err := ParseV3(data)
	if err != nil {
		return err
	}
	err = applyTransformations(ctx, cfg, func(ctx context.Context, t transforms.T, node yaml.Node) (err error) {
		doc, err = t.Transform(doc)
		return
	})
	if err != nil {
		return err
	}
	return formatAndWriteV3(ctx, filename, doc, fv.OutputFlags, asYAML, fv.Validate)
}

type applyFunc func(ctx context.Context, t transforms.T, node yaml.Node) error

func applyTransformations(ctx context.Context, cfg transforms.Config, fn applyFunc) error {
	for i, n := range cfg.Transforms {
		t := transforms.Get(n)
		if t == nil {
			return fmt.Errorf("transform %v is not installed: must be one of: %v", n, strings.Join(transforms.List(), ", "))
		}
		node := cfg.Configs[i]
		if err := fn(ctx, t, node); err != nil {
			return err
		}
	}
	return nil
}
