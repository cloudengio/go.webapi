// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms_test

import (
	"testing"

	"cloudeng.io/webapi/openapi/transforms"
)

const replacementConfig = `configs:
  - replacements:
    - path: [components, schemas, API, properties, ignoreEg, allOf]
      replacement:
        example: new example
        type: string
`

func TestReplacement(t *testing.T) {
	doc, cfg := loadForTest("allof-eg.yaml", replacementConfig)
	tr := transforms.Get("replacements")
	if err := cfg.ConfigureAll(); err != nil {
		t.Fatal(err)
	}
	doc, err := tr.Transform(doc)
	if err != nil {
		t.Fatal(err)
	}
	txt := asYAML(t, doc)
	contains(t, 4, txt, `
API:
  properties:
    ignoreEg:
      example: new example
      type: string
API2:
`)
}
