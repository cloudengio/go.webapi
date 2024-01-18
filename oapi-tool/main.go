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
    summary: display the element at paths in an openapi v3 specification. Paths are colon separated, eg. ':components:schemas:MySchema'.
    arguments:
      - path
      - ...
`

var cmdSet *subcmd.CommandSetYAML

func init() {
	cmdSet = subcmd.MustFromYAML(spec)
	cmdSet.Set("download").MustRunner(downloadCmd, &DownloadFlags{})
	cmdSet.Set("format").MustRunner(formatCmd, &FormatFlags{})
	cmdSet.Set("transform").MustRunner(transformCmd, &TransformFlags{})
	cmdSet.Set("validate").MustRunner(validateCmd, &struct{}{})
	cmdSet.Set("convert").MustRunner(convertCmd, &ConvertFlags{})
	cmdSet.Set("inspect").MustRunner(inspectCmd, &InspectFlags{})
}

func main() {
	cmdSet.MustDispatch(context.Background())
}
