// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
)

func validateCmd(ctx context.Context, values any, args []string) error {
	filename := args[0]
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(filename)
	if err != nil {
		return err
	}
	return doc.Validate(ctx)
}
