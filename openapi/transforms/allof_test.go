// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms_test

import (
	"testing"

	"cloudeng.io/webapi/openapi/transforms"
)

const allOfConfig = `configs:
  - allOf:
    - path: [components, schemas, API, properties, ignoreEg]
      ignoreNonType: true
    - path: [components, schemas, API2, properties, promoteEg]
      promoteNonType: [readOnly, description]
    - path: [components, schemas, API3, properties, promoteEg]
      promoteNonType: [properties]
    - path: [components, schemas, MergeEg, properties, API]
      mergeNonType: [properties]
  `

func TestAllOf(t *testing.T) {
	doc, cfg := loadForTest("allof-eg.yaml", allOfConfig)
	tr := transforms.Get("allOf")
	if err := cfg.ConfigureAll(); err != nil {
		t.Fatal(err)
	}
	doc, err := tr.Transform(doc)
	if err != nil {
		t.Fatal(err)
	}
	// ignore & promote readonly
	// merge properties - which is possibly not needed.
	txt := asYAML(t, doc)
	contains(t, 4, txt, `
API:
  properties:
    ignoreEg:
      allOf:
        - type: object
      type: object
API2:
`)
	contains(t, 4, txt, `
API2:
  properties:
    promoteEg:
      allOf:
        - type: object
      description: something
      readOnly: true
      type: object
API3:
`)
	contains(t, 4, txt, `
API3:
  properties:
    promoteEg:
      allOf:
        - type: object
      properties:
        egURL:
          description: a URL
      type: object
Base:
`)
	contains(t, 4, txt, `
MergeEg:
  properties:
    API:
      allOf:
        - properties:
            egURL:
              description: describe the previous type
          type: object
      type: object
`)
}
