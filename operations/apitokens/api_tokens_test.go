// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens_test

import (
	"context"
	"testing"

	"cloudeng.io/webapi/operations/apitokens"
)

func TestAPITokens(t *testing.T) {
	ctx := context.Background()
	has := func(n, v string) {
		token, ok := apitokens.TokenFromContext(ctx, n)
		if got, want := ok, true; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token, v; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	notHas := func(n string) {
		token, ok := apitokens.TokenFromContext(ctx, n)
		if got, want := ok, false; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := token, ""; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	ctx = apitokens.ContextWithToken(ctx, "n1", "v1")
	has("n1", "v1")
	notHas("n2")
	ctx = apitokens.ContextWithToken(ctx, "n2", "v2")
	has("n1", "v1")
	has("n2", "v2")
	notHas("n3")
}
