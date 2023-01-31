// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

/*
import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type TransformFlags struct {
	OutputFlags
	Config   string `subcmd:"config,transform.yaml,yaml configuration for the transformations to be applied"`
	Describe bool   `subcmd:"describe,,describe all configured transformations"`
}

func transformCmd(ctx context.Context, values any, args []string) error {
	fv := values.(*TransformFlags)
	cfg, err := loadConfig(fv.Config)
	if err != nil {
		return err
	}

	if fv.Describe {
		return applyTransformations(ctx, cfg, func(ctx context.Context, t Transformer, node yaml.Node) error {
			out := t.Describe(node)
			fmt.Println(out)
			return nil
		})
	}

	err = applyTransformations(ctx, cfg, func(ctx context.Context, t Transformer, node yaml.Node) error {
		return t.Configure(node)
	})
	if err != nil {
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
	if doc != nil && IsV2(doc) {
		doc, err := ParseV2(data)
		if err != nil {
			return err
		}
		err = applyTransformations(ctx, cfg, func(ctx context.Context, t Transformer, node yaml.Node) (err error) {
			doc, err = t.TransformV2(doc)
			return
		})
		if err != nil {
			return err
		}
		return formatAndWriteV2(filename, doc, fv.OutputFlags)
	}
	if err != nil {
		return err
	}
	err = applyTransformations(ctx, cfg, func(ctx context.Context, t Transformer, node yaml.Node) (err error) {
		doc, err = t.TransformV3(doc)
		return
	})
	if err != nil {
		return err
	}
	return formatAndWriteV3(filename, doc, fv.OutputFlags, asYAML)
}

type applyFunc func(ctx context.Context, t Transformer, node yaml.Node) error

func applyTransformations(ctx context.Context, cfg *TransformConfig, fn applyFunc) error {
	for i, n := range cfg.Names {
		t, ok := installed[n]
		if !ok {
			return fmt.Errorf("transform %v is not installed: must be one of: %v", n, strings.Join(listInstalled(), ", "))
		}
		node := cfg.Transforms[i]
		if err := fn(ctx, t, node); err != nil {
			return err
		}
	}
	return nil
}
*/
