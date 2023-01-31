// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

import (
	"strings"

	"cloudeng.io/text/linewrap"
	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&replacementTransformer{})
}

type replacements struct {
	Path        []string `yaml:",flow"`
	Replacement yaml.Node
	replacement map[string]any
}

type replacementTransformer struct {
	ReplacementRules []replacements `yaml:"replacements"`
}

func (t *replacementTransformer) Name() string {
	return "replacements"
}

func (t *replacementTransformer) Configure(node yaml.Node) error {
	var rw []replacements
	if err := node.Decode(&rw); err != nil {
		return err
	}
	t.ReplacementRules = rw
	for i, r := range rw {
		if err := r.Replacement.Decode(&rw[i].replacement); err != nil {
			return err
		}
	}
	return nil
}

func (t *replacementTransformer) Describe(node yaml.Node) string {
	out := &strings.Builder{}
	out.WriteString(linewrap.Block(0, 80, `
The replacement transform handles cases where entire blocks of
the specification need to be replaced with new ones.`))
	tmp := &replacementTransformer{}
	node.Decode(tmp)
	out.WriteString("\noptions:\n")
	out.WriteString(formatYAML(2, tmp))
	return out.String()
}

func (t *replacementTransformer) Transform(doc *openapi3.T) (*openapi3.T, error) {
	walker := openapi.NewWalker(t.visitor)
	return doc, walker.Walk(doc)
}

func (t *replacementTransformer) visitor(path []string, parent, node any) (bool, error) {
	for _, repl := range t.ReplacementRules {
		if !match(path, repl.Path) {
			continue
		}
		pmap := jsonMap(parent)
		pmap[path[len(path)-1]] = nil
		for k, v := range repl.replacement {
			pmap[k] = v
		}
		if err := marshalMap(pmap, parent); err != nil {
			return false, err
		}
	}
	return true, nil
}
