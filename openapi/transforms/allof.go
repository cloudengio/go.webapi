// Copyright 2022 cloudeng llc. All rights reserved.
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
	Register(&allOfTransformer{})
}

type allOf struct {
	Path           []string `yaml:",flow"`
	IgnoreNonType  bool     `yaml:"ignoreNonType"`
	PromoteNonType []string `yaml:"promoteNonType,flow"`
	MergeNonType   []string `yaml:"mergeNonType,flow"`
}

type allOfTransformer struct {
	AllOfRules []allOf `yaml:"allOf"`
}

func (t *allOfTransformer) Name() string {
	return "allOf"
}

func (t *allOfTransformer) Configure(node yaml.Node) error {
	var ao []allOf
	if err := node.Decode(&ao); err != nil {
		return err
	}
	t.AllOfRules = ao
	return nil
}

func (t *allOfTransformer) Describe(node yaml.Node) string {
	out := &strings.Builder{}
	out.WriteString(linewrap.Block(0, 80, `
The allOf transform handles cases where the allOf array members are incorrectly
specified. In particular it currently allows for ignoring, promoting or merging
allOf entries that are not themselves schemas with type information.`))
	tmp := &allOfTransformer{}
	if err := node.Decode(tmp); err != nil {
		fmt.Fprintf(out, "error: %v", err)
		return out.String()
	}
	out.WriteString("\noptions:\n")
	out.WriteString(formatYAML(2, tmp))
	return out.String()
}

func (t *allOfTransformer) Transform(doc *openapi3.T) (*openapi3.T, error) {
	walker := openapi.NewWalker(t.visitor)
	return doc, walker.Walk(doc)
}

func hasSchema(s *openapi3.SchemaRef) bool {
	if len(s.Ref) > 0 {
		return true
	}
	v := s.Value
	if len(v.Type) > 0 || len(v.AllOf) > 0 || len(v.OneOf) > 0 || len(v.AnyOf) > 0 || len(v.Enum) > 0 {
		return true
	}
	for _, p := range v.Properties {
		if hasSchema(p) {
			return true
		}
	}
	return false
}

func containsNonType(srefs openapi3.SchemaRefs) bool {
	for _, sr := range srefs {
		if !hasSchema(sr) {
			return true
		}
	}
	return false
}

func handleMerge(a, b any, fields []string) (any, error) {
	am, bm := jsonMap(a), jsonMap(b)
	for _, field := range fields {
		if v, ok := bm[field]; ok {
			am[field] = v
		}
	}
	c := a
	err := marshalMap(am, c)
	return c, err
}

func (t *allOfTransformer) handleTransformation(r allOf, schema *openapi3.SchemaRef) error {
	na := []*openapi3.SchemaRef{}
	var prev *openapi3.SchemaRef
	for i, s := range schema.Value.AllOf {
		if hasSchema(s) {
			prev = s
			na = append(na, s)
			continue
		}
		switch {
		case r.IgnoreNonType:
			continue
		case len(r.MergeNonType) > 0:
			if prev == nil {
				return fmt.Errorf("allOf entry: %v cannot be merged since there is previous schema with a type to merge it with", i)
			}
			fmt.Printf("merging.....\n")
			n, err := handleMerge(prev.Value, s.Value, r.MergeNonType)
			if err != nil {
				return err
			}
			prev.Value = n.(*openapi3.Schema)
		case len(r.PromoteNonType) > 0:
			n, err := handleMerge(schema.Value, s.Value, r.PromoteNonType)
			if err != nil {
				return err
			}
			schema.Value = n.(*openapi3.Schema)
		}
	}
	schema.Value.AllOf = na
	return nil
}

func (t *allOfTransformer) visitor(path []string, parent, node any) (bool, error) {
	schema, ok := node.(*openapi3.SchemaRef)
	if !ok {
		return true, nil
	}
	if len(schema.Value.AllOf) == 0 {
		return true, nil
	}
	if !containsNonType(schema.Value.AllOf) {
		return true, nil
	}
	for _, r := range t.AllOfRules {
		if !match(path, r.Path) {
			continue
		}
		if err := t.handleTransformation(r, schema); err != nil {
			return false, fmt.Errorf("%v: %v", strings.Join(path, ":"), err)
		}
	}
	return true, nil
}
