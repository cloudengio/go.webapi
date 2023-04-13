// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocolsiocmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cloudeng.io/file/content"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/protocolsio/protocolsiosdk"
)

// ReadDownload reads the downloaded object from the specified path.
func ReadDownload(path string) (*content.Object[protocolsiosdk.ProtocolPayload, operations.Response], error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	obj := new(content.Object[protocolsiosdk.ProtocolPayload, operations.Response])
	if err := obj.Decode(buf); err != nil {
		return nil, err
	}
	return obj, nil
}

// WriteDownload writes the downloaded object to the specified path.
func WriteDownload(path string, obj content.Object[protocolsiosdk.ProtocolPayload, operations.Response]) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := obj.WriteObjectFile(path, content.GOBObjectEncoding, content.GOBObjectEncoding); err != nil {
		fmt.Printf("failed to write: %v as %v: %v\n", obj.Value.Protocol.ID, path, err)
	}
	return nil
}
