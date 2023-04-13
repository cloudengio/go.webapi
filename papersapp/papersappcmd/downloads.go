// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersappcmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"cloudeng.io/file/content"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/papersapp"
)

// ReadItemDownload reads the downloaded object from the specified path.
func ReadItemDownload(path string) (*content.Object[papersapp.Item, operations.Response], error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	obj := new(content.Object[papersapp.Item, operations.Response])
	if err := obj.Decode(buf); err != nil {
		return nil, err
	}
	return obj, nil
}

var written int64

// WriteItemDownload writes the downloaded object to the specified path.
func WriteDownload[T any](path string, obj content.Object[T, operations.Response]) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := obj.WriteObjectFile(path, content.JSONObjectEncoding, content.GOBObjectEncoding); err != nil {
		fmt.Printf("failed to write: %T %v as %v: %v\n", obj.Value, obj.Value, path, err)
	}
	if w := atomic.AddInt64(&written, 1); w%100 == 0 {
		log.Printf("files written: %v", w)
	}
	return nil
}
