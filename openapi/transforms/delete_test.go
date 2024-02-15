// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms_test

import (
	"testing"

	"cloudeng.io/webapi/openapi/transforms"
)

const deletionConfig = `configs:
  - deletions:
    - path: [components, schemas, api, properties, id]
      field: example
    - path: [components, schemas, api, properties, end]
      field: type
`

func TestDeletion(t *testing.T) {
	doc, cfg := loadForTest("rewrite-eg.yaml", deletionConfig)
	tr := transforms.Get("deletions")
	if err := cfg.ConfigureAll(); err != nil {
		t.Fatal(err)
	}
	doc, err := tr.Transform(doc)
	if err != nil {
		t.Fatal(err)
	}
	txt := asYAML(t, doc)
	contains(t, 8, txt, `
end:
  example: example_replacement
id:
  type: string
name:
`)
}
