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
	Register(&discriminatorTransformer{})
}

type discriminatorRule struct {
	PathPrefix     []string `yaml:"pathPrefix,flow"`
	CreateProperty bool     `yaml:"createProperty"`
	CreateRequired bool     `yaml:"createRequired"`
}

type discriminatorTransformer struct {
	DiscriminatorRules []discriminatorRule `yaml:"discriminator"`
}

func (t *discriminatorTransformer) Name() string {
	return "discriminator"
}

func (t *discriminatorTransformer) Configure(node yaml.Node) error {
	var dr []discriminatorRule
	if err := node.Decode(&dr); err != nil {
		return err
	}
	t.DiscriminatorRules = dr
	return nil
}

func (t *discriminatorTransformer) Describe(node yaml.Node) string {
	out := &strings.Builder{}
	out.WriteString(linewrap.Block(0, 80, `
The discriminator transform handles cases where a oneOf or anyOf specification is incomplete. For example if its discriminator is not listed as a property. This is typically required by some code generators.`))
	tmp := &allOfTransformer{}
	node.Decode(tmp)
	out.WriteString("\noptions:\n")
	out.WriteString(formatYAML(2, tmp))
	return out.String()
}

func (t *discriminatorTransformer) Transform(doc *openapi3.T) (*openapi3.T, error) {
	walker := openapi.NewWalker(t.visitor)
	return doc, walker.Walk(doc)
}

func (t *discriminatorTransformer) handleProperty(dr discriminatorRule, schema *openapi3.Schema) {
	if !dr.CreateProperty {
		return
	}
	discName := schema.Discriminator.PropertyName
	for m := range schema.Properties {
		if m == discName {
			return
		}
	}
	if schema.Properties == nil {
		schema.Properties = openapi3.Schemas{}
	}
	schema.Properties[discName] = &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: "string",
		},
	}
}

func (t *discriminatorTransformer) handleRequired(dr discriminatorRule, schema *openapi3.Schema) {
	if !dr.CreateRequired {
		return
	}
	discName := schema.Discriminator.PropertyName
	for _, m := range schema.Required {
		if m == discName {
			return
		}
	}
	schema.Required = append(schema.Required, discName)
}

func (t *discriminatorTransformer) visitor(path []string, parent, node any) (bool, error) {
	schema, ok := node.(*openapi3.SchemaRef)
	if !ok {
		return true, nil
	}
	if schema.Value.Discriminator == nil {
		return true, nil
	}

	for _, dr := range t.DiscriminatorRules {
		t.handleProperty(dr, schema.Value)
		t.handleRequired(dr, schema.Value)
	}
	return true, nil
}
