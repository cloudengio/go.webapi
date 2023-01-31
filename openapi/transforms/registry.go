// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

import (
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// T represents a 'Transformer' that can be used to perform structured
// edits/transforms on an openapi 3 specification.
type T interface {
	Name() string
	Describe(node yaml.Node) string
	Configure(node yaml.Node) error
	Transform(*openapi3.T) (*openapi3.T, error)
}

var installed = map[string]T{}

// Register registers a transformer and make it available to clients
// of this package.
func Register(t T) {
	installed[t.Name()] = t
}

// List returns a list of all available transformers.
func List() []string {
	var r []string
	for k := range installed {
		r = append(r, k)
	}
	return r
}

// Get returns the transformer, if any, for the requested name. It returns
// nil if no transformer with that name has been registered.
func Get(name string) T {
	return installed[name]
}
