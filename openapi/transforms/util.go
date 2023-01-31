// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func formatYAML(indent int, v any) string {
	out := &strings.Builder{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(indent)
	if err := enc.Encode(v); err != nil {
		return ""
	}
	return out.String()
}

// Replacement represents a replacement string of the
// form /<match-re>/<replacement>/
type Replacement struct {
	match   *regexp.Regexp
	replace string
}

// Match applies regexp.MatchString.
func (sr Replacement) MatchString(input string) bool {
	return sr.match.MatchString(input)
}

// ReplaceAllString(input string) applies regexp.ReplaceAllString.
func (sr Replacement) ReplaceAllString(input string) string {
	return sr.match.ReplaceAllString(input, sr.replace)
}

// NewReplacement accepts a string of the form /<match-re>/<replacement>/
// to create a Replacement that will apply
// <match-re.ReplaceAllString(<replace>).
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
