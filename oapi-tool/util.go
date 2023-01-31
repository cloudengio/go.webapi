// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func YAMLOrJSON(path string) (bool, error) {
	switch {
	case strings.HasSuffix(path, ".json"):
		return false, nil
	case strings.HasSuffix(path, ".yaml"):
		fallthrough
	case strings.HasSuffix(path, ".yml"):
		return true, nil
	}
	return false, fmt.Errorf("unsupported file type (needs to .json, .yaml or .yml: %v", path)
}

func ParseV3(data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func IsV2(doc *openapi3.T) bool {
	return doc.OpenAPI[0] == '2'
}

func FormatJSON(t any) string {
	buf, _ := json.MarshalIndent(t, "", " ")
	return string(buf)
}

func FormatYAML(indent int, v any) string {
	out := &strings.Builder{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(indent)
	if err := enc.Encode(v); err != nil {
		return ""
	}
	return out.String()
}

func OutputFormat(filename string, convertToYAML bool) (bool, error) {
	if convertToYAML {
		return true, nil
	}
	return YAMLOrJSON(filename)
}

type Replacement struct {
	match   *regexp.Regexp
	replace string
}

func (sr Replacement) Match(input string) bool {
	return sr.match.MatchString(input)
}

func (sr Replacement) Replace(input string) string {
	return sr.match.ReplaceAllString(input, sr.replace)
}

func NewReplacement(s string) (Replacement, error) {
	var sr Replacement
	var parts []string
	for _, p := range strings.Split(s, "/") {
		if len(p) > 0 {
			parts = append(parts, p)
		}
	}
	if len(parts) != 2 {
		return sr, fmt.Errorf("%q is not in /<match>/<replace>/ form", s)
	}
	m, err := regexp.Compile(parts[0])
	if err != nil {
		return sr, err
	}
	sr.match = m
	sr.replace = parts[1]
	return sr, nil
}

/*
type rewriteRule struct {
	Rewrite string `yaml:"rewrite"`
	re      SedReplacement
}

func (rw *rewriteRule) configure() error {
	var err error
	var errs errors.M
	if len(rw.Rewrite) > 0 {
		rw.re, err = NewSedReplacement(rw.Rewrite)
		errs.Append(err)
	}
	return errs.Err()
}
*/
