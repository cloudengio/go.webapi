// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package openapi_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi3"
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

type testVisitor struct {
	paths []string
}

func (t *testVisitor) visitor(path []string, parent, node any) (bool, error) {
	t.paths = append(t.paths, strings.Join(path, ":"))
	return true, nil
}

func TestWalk(t *testing.T) {
	doc := loadYAML("benchling.yaml")
	v := &testVisitor{paths: []string{}}
	wk := openapi.NewWalker(v.visitor)
	if err := wk.Walk(doc); err != nil {
		t.Fatal(err)
	}
	sort.Strings(v.paths)
	buf, err := os.ReadFile(filepath.Join("testdata", "benchling.paths.txt"))
	if err != nil {
		t.Fatal(err)
	}
	v.paths = append(v.paths, "") // trailing newline in the test data
	lines := strings.Split(string(buf), "\n")
	if got, want := len(v.paths), len(lines); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, line := range lines {
		if got, want := v.paths[i], line; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	v.paths = nil
	wk = openapi.NewWalker(v.visitor, openapi.WalkerVisitPrefix("components", "schemas", "AaSequence", "properties", "webURL"))
	if err := wk.Walk(doc); err != nil {
		t.Fatal(err)
	}

	if got, want := v.paths, []string{"components:schemas:AaSequence:properties:webURL"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
