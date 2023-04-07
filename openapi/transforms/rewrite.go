// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

import (
	"encoding/json"
	"fmt"
	"strings"

	"cloudeng.io/text/linewrap"
	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&rewriteTransformer{})
}

type rewriteRule struct {
	Path    []string `yaml:"path,flow"`
	Rewrite string
	Replace string
	repl    Replacement
}

type rewriteTransformer struct {
	Rewrites []rewriteRule
}

func (t *rewriteTransformer) configure(rewrites []rewriteRule) ([]rewriteRule, error) {
	for i, rw := range rewrites {
		repl, err := NewReplacement(rw.Rewrite)
		if err != nil {
			return nil, err
		}
		rewrites[i].repl = repl
	}
	return rewrites, nil
}

func (t *rewriteTransformer) Name() string {
	return "rewrites"
}

func (t *rewriteTransformer) Configure(node yaml.Node) error {
	var rw []rewriteRule
	if err := node.Decode(&rw); err != nil {
		return err
	}
	rw, err := t.configure(rw)
	t.Rewrites = rw
	return err
}

func (t *rewriteTransformer) Describe(node yaml.Node) string {
	out := &strings.Builder{}
	fmt.Fprintf(out, "%s", linewrap.Block(0, 80, `
The rewrites transform rewriting of fields using expressions of the form "/regexp/replacement/"
`))
	tmp := &rewriteTransformer{}
	if err := node.Decode(tmp); err != nil {
		fmt.Fprintf(out, "error: %v", err)
		return out.String()
	}
	out.WriteString("\noptions:\n")
	out.WriteString(formatYAML(2, tmp))
	return out.String()
}

func (t *rewriteTransformer) Transform(doc *openapi3.T) (*openapi3.T, error) {
	walker := openapi.NewWalker(t.visitor)
	return doc, walker.Walk(doc)
}

func jsonMap(v any) map[string]any {
	var r map[string]any
	buf, _ := json.Marshal(v)
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil
	}
	return r
}

func marshalMap(in map[string]any, out any) error {
	buf, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, out)
}

func (t *rewriteTransformer) visitor(path []string, _, node any) (bool, error) {
	for _, rw := range t.Rewrites {
		if len(rw.Replace) == 0 {
			continue
		}
		if !match(path, rw.Path) {
			continue
		}
		fields := jsonMap(node)
		ov, ok := fields[rw.Replace].(string)
		if !ok {
			return false, fmt.Errorf("%v:%v is not a string", strings.Join(path, ":"), rw.Replace)
		}
		if !rw.repl.MatchString(ov) {
			continue
		}
		fields[rw.Replace] = rw.repl.ReplaceAllString(ov)
		if err := marshalMap(fields, node); err != nil {
			fields[rw.Replace] = ov
			return false, fmt.Errorf("%v:%v failed to update new value: %v", strings.Join(path, ":"), rw.Replace, err)
		}
	}
	return true, nil
}
