// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"

	"cloudeng.io/cmdutil/subcmd"
)

const spec = `
name: openapi
summary: command line for manipulating openapi/swagger specifications.
commands:
  - name: download
    summary: download a specification.
    arguments:
      - url
  - name: format
    summary: format a specification.
    arguments:
      - filename
  - name: transform
    summary: modify an openapi v3 specification using a set of built in transformations
             and rewrites.
    arguments:
      - filename
  - name: validate
    summary: validate an openapi v3 specification.
    arguments:
      - filename
  - name: convert
    summary: convert an openapi v2 specification to v3.
    arguments:
      - filename
  - name: inspect
    summary: display the element at paths in an openapi v3 specification. Paths are / separated, eg. /components/schemas/MySchema.
    arguments:
      - path
      - ...
`

var cmdSet *subcmd.CommandSetYAML

func init() {
	cmdSet = subcmd.MustFromYAML(spec)
	cmdSet.Set("download").MustRunnerAndFlags(downloadCmd,
		subcmd.MustRegisteredFlagSet(&DownloadFlags{}))
	cmdSet.Set("format").MustRunnerAndFlags(formatCmd,
		subcmd.MustRegisteredFlagSet(&FormatFlags{}))
	cmdSet.Set("transform").MustRunnerAndFlags(transformCmd,
		subcmd.MustRegisteredFlagSet(&TransformFlags{}))
	cmdSet.Set("validate").MustRunnerAndFlags(validateCmd,
		subcmd.MustRegisteredFlagSet(&struct{}{}))
	cmdSet.Set("convert").MustRunnerAndFlags(convertCmd,
		subcmd.MustRegisteredFlagSet(&ConvertFlags{}))
	cmdSet.Set("inspect").MustRunnerAndFlags(inspectCmd,
		subcmd.MustRegisteredFlagSet(&InspectFlags{}))
}

func main() {
	cmdSet.MustDispatch(context.Background())
}
