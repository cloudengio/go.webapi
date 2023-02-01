// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package openapi_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func load(filename string) *openapi3.T {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}
	return doc
}

func loadV2(filename string) *openapi2.T {
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}
	var doc openapi2.T
	if err := json.Unmarshal(data, &doc); err != nil {
		panic(err)
	}
	return &doc
}

func TestFormat(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		filename  string
		validates bool
	}{
		{"api.yaml", true},
		{"petstore-expanded.yaml", false},
	} {
		// Make sure that anything we format can be parsed back
		// correctly.
		filename := tc.filename
		doc := load(filename)
		buf, err := openapi.FormatV3(doc, true)
		if err != nil {
			t.Fatalf("%v: %v", filename, err)
		}
		var tmp any
		if err := yaml.Unmarshal(buf, &tmp); err != nil {
			t.Fatalf("%v: %v", filename, err)
		}
		loader := openapi3.NewLoader()
		ndoc, err := loader.LoadFromData(buf)
		if err != nil {
			t.Fatalf("%v: %v", filename, err)
		}
		if tc.validates {
			// Make sure that files are known to be valid are not
			// messed up by (re)formatting.
			if err := ndoc.Validate(ctx); err != nil {
				t.Errorf("%v: %v", filename, err)
			}
		}
	}
}

func TestFormatStyle(t *testing.T) {
	ctx := context.Background()
	doc2 := loadV2("v2swagger.json")
	doc3, err := openapi2conv.ToV3(doc2)
	if err != nil {
		t.Fatal(err)
	}

	// As JSON
	buf, err := openapi.FormatV3(doc3, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(buf), `{"components":{"schemas":{"ApiResponse`) {
		t.Fatalf("doesn't look like JSON: %s", buf)
	}

	loader := openapi3.NewLoader()
	ndocJSON, err := loader.LoadFromData(buf)
	if err != nil {
		t.Fatal(err)
	}
	if err := ndocJSON.Validate(ctx); err != nil {
		t.Fatal(err)
	}

	// As YAML
	buf, err = openapi.FormatV3(doc3, true)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(buf), `components:`) {
		t.Fatalf("doesn't look like YAML: %s", buf)
	}

	ndocYAML, err := loader.LoadFromData(buf)
	if err != nil {
		t.Fatal(err)
	}
	if err := ndocYAML.Validate(ctx); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ndocJSON, ndocYAML) {
		t.Fatal(err)
	}

}
