// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

/*
import (
	"strings"

	"cloudeng.io/errors"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func init() {
	RegisterTransform("dates", &fixDates{})
}

type fixDates struct {
	Date     rewriteRule `yaml:"date"`
	DateTime rewriteRule `yaml:"date-time"`
}

func (t *fixDates) Configure(node yaml.Node) error {
	if err := node.Decode(t); err != nil {
		return err
	}
	var errs errors.M
	errs.Append((&t.Date).configure())
	return errs.Err()
}

func (t *fixDates) Describe(node yaml.Node) string {
	out := &strings.Builder{}

	return out.String()
}

func (t *fixDates) TransformV2(doc *openapi2.T) (*openapi2.T, error) {
	return nil, ErrTransformNotImplementedForV2
}

func (t *fixDates) TransformV3(doc *openapi3.T) (*openapi3.T, error) {
	visitSchemas(doc, t.visitor)
	return doc, nil
}

func (t *fixDates) visitor(sr *openapi3.SchemaRef) {
	v := sr.Value
	if len(v.Type) == 0 {
		return
	}
	var egType string
	var eg any
	if v.Type == openapi3.TypeString {
		//		fmt.Printf("[%v].... %v .... \n", sr.Value.Type, FormatYAML(2, sr))
		switch v.Format {
		case "date", "date-time":
			egType = v.Format
			eg = v.Example
		default:
			return
		}
	}
	//	fmt.Printf("... %v\n", v.Type)
	//	if v.Type == "enum" {
	//		for _, e := range v.Enum {
	//			fmt.Printf(".... %T\n", e)
	//		}
	//	}
	_, _ = eg, egType
}
*/
