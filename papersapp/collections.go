// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package papersapp

import (
	"context"

	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/papersapp/papersappsdk"
)

func ListCollections(ctx context.Context, serviceURL string, opts ...operations.Option) ([]*papersappsdk.Collection, error) {
	ep := operations.NewEndpoint[papersappsdk.Collections](opts...)
	cols, _, _, err := ep.Get(ctx, serviceURL+"/collections")
	if err != nil {
		return nil, err
	}
	return cols.Collections, nil
}
