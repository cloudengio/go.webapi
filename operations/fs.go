// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"cloudeng.io/file/content"
	"cloudeng.io/file/filewalk"
)

// FS defines a filesystem interface to be broadly used by webapi packages
// and clients. It is defined in operations for convenience.
type FS interface {
	content.FS
	filewalk.FS
}
