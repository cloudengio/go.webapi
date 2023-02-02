// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import "net/http"

type Stream interface {
	Next(*http.Response) (*http.Request, bool, error)
}
