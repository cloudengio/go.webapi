// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

import (
	"fmt"
	"strings"

	"cloudeng.io/text/linewrap"
	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&deletionTransformer{})
}

type deletionRule struct {
	Path  []string `yaml:"path,flow"`
	Field string   `yaml:"field"`
}

type deletionTransformer struct {
	Deletions []deletionRule
}

func (t *deletionTransformer) Name() string {
	return "deletions"
}

func (t *deletionTransformer) Configure(node yaml.Node) error {
	return node.Decode(&t.Deletions)
}

func (t *deletionTransformer) Describe(node yaml.Node) string {
	out := &strings.Builder{}
	fmt.Fprintf(out, "%s", linewrap.Block(0, 80, `
The deletion transform deletes the specified paths from the openapi specification."
`))
	tmp := &deletionTransformer{}
	if err := node.Decode(tmp); err != nil {
		fmt.Fprintf(out, "error: %v", err)
		return out.String()
	}
	out.WriteString("\noptions:\n")
	out.WriteString(formatYAML(2, tmp))
	return out.String()
}

func (t *deletionTransformer) Transform(doc *openapi3.T) (*openapi3.T, error) {
	walker := openapi.NewWalker(t.visitor)
	return doc, walker.Walk(doc)
}

func (t *deletionTransformer) visitor(path []string, _, node any) (bool, error) {
	for _, rw := range t.Deletions {
		if !match(path, rw.Path) {
			continue
		}
		fields := jsonMap(node)
		ov := fields[rw.Field]
		delete(fields, rw.Field)
		if err := marshalMap(fields, node); err != nil {
			fields[rw.Field] = ov
			return false, fmt.Errorf("%v:%v failed to delete field: %v", strings.Join(path, ":"), rw.Field, err)
		}
	}
	return true, nil
}
