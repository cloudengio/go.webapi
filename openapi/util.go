// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package openapi

import (
	"bytes"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func FormatV3(doc *openapi3.T, isYAML bool) ([]byte, error) {
	// roundtrip to/from json to ensure corner cases are handled
	// correctly. See: http://web.archive.org/web/20190603050330/http://ghodss.com/2014/the-right-way-to-handle-yaml-in-golang/
	data, err := doc.MarshalJSON()
	if !isYAML {
		return data, err
	}
	var tmp any
	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	err = enc.Encode(tmp)
	return out.Bytes(), err
}

func AsYAML(indent int, doc any) (string, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	var tmp any
	if err := json.Unmarshal(data, &tmp); err != nil {
		return "", err
	}
	out := &bytes.Buffer{}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(indent)
	err = enc.Encode(tmp)
	return out.String(), err
}
