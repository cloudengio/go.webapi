// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"os"

	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type OutputFlags struct {
	Output        string `subcmd:"output,,'filename to save the formatted file as, if not set, the contents will be written to stdout'"`
	OverWrite     bool   `subcmd:"over-write,false,set to true to over-write the formated file in place"`
	ConvertToYAML bool   `subcmd:"yaml,false,the output will be in YAML format regardless of the input format"`
}

type FormatFlags struct {
	OutputFlags
	Validate bool `subcmd:"validate,true,validate the formatted specification"`
}

func formatCmd(ctx context.Context, values any, args []string) error {
	fv := values.(*FormatFlags)
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
	return formatAndWriteV3(ctx, filename, doc, fv.OutputFlags, asYAML, fv.Validate)
}

func formatAndWriteV3(ctx context.Context, filename string, doc *openapi3.T, fv OutputFlags, isYAML bool, validate bool) error {
	data, err := openapi.FormatV3(doc, isYAML)
	if err != nil {
		return err
	}
	if validate {
		fdoc, err := ParseV3(data)
		if err != nil {
			return err
		}
		if err := fdoc.Validate(ctx); err != nil {
			return err
		}
	}
	return writeFormatted(filename, data, fv)
}

func writeFormatted(filename string, data []byte, fv OutputFlags) error {
	if len(fv.Output) == 0 {
		if fv.OverWrite {
			return os.WriteFile(filename, data, 0660)
		}
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(fv.Output, data, 0660)
}
