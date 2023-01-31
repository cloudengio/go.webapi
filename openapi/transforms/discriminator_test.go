// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms_test

import (
	"testing"

	"cloudeng.io/webapi/openapi/transforms"
)

const discrimatorfConfig = `configs:
  - discriminator:
    - pathPrefix: [components, schemas]
      createProperty: true
      createRequired: true

`

func TestDiscriminatorf(t *testing.T) {
	doc, cfg := loadForTest("discriminator-eg.yaml", discrimatorfConfig)
	tr := transforms.Get("discriminator")
	if err := cfg.ConfigureAll(); err != nil {
		t.Fatal(err)
	}
	doc, err := tr.Transform(doc)
	if err != nil {
		t.Fatal(err)
	}
	txt := asYAML(t, doc)
	contains(t, 4, txt, `
Pet:
  type: object
  required:
    - pet_type
  properties:
    pet_type:
      type: string
    something_else:
      type: string
  discriminator:
    propertyName: pet_type
`)
}
