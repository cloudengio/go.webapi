// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"cloudeng.io/webapi/openapi/transforms"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func loadYAML(filename string) *openapi3.T {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}
	return doc
}

func loadForTest(filename string, config string) (*openapi3.T, transforms.Config) {
	doc := loadYAML(filename)
	cfg, err := transforms.ParseConfig([]byte(config))
	if err != nil {
		panic(err)
	}
	return doc, cfg
}

const rewriteConfig = `configs:
  - rewrites:
    - path: [components, schemas, api, properties, id]
      rewrite: "/^example_replacement$/something-new/"
      replace: example
    - path: [components, schemas, api, properties, end]
      rewrite: "/^integer_error$/integer/"
      replace: type
 `

func asYAML(t *testing.T, v any) string {
	out := &strings.Builder{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(1)
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func trimIndent(indent int, s string) string {
	spaces := strings.Repeat(" ", indent)
	out := strings.Builder{}
	for _, l := range strings.Split(s, "\n") {
		out.WriteString(strings.TrimPrefix(l, spaces))
		out.WriteRune('\n')
	}
	return strings.TrimSpace(out.String())
}

func contains(t *testing.T, indent int, got, want string) {
	got = trimIndent(indent, got)
	if !strings.Contains(got, want) {
		_, file, line, _ := runtime.Caller(1)
		file = filepath.Base(file)
		t.Errorf("%s:%v: %v\n\ndoes not contain\n\n%v", file, line, got, want)
	}
}

func TestRewrite(t *testing.T) {
	doc, cfg := loadForTest("rewrite-eg.yaml", rewriteConfig)
	tr := transforms.Get("rewrites")
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
  type: integer
  example: example_replacement
id:
  type: string
  example: something-new
name:
  type: string
  maxLength: 255
start:
  type: integer
  example: example_replacement
`)
}
