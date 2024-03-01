// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens_test

import (
	"context"
	"os"
	"testing"

	"cloudeng.io/webapi/operations/apitokens"
)

func TestTokenReaders(t *testing.T) {
	ctx := context.Background()
	os.Setenv("something-unlikely", "env-secret")
	for _, test := range []struct {
		text   string
		scheme string
		name   string
		value  string
	}{
		{"file://testdata/example.tok", "file", "testdata/example.tok", "example-token\n"},
		{"env://something-unlikely", "env", "something-unlikely", "env-secret"},
		{"literal://right-here", "literal", "right-here", "right-here"},
	} {
		tok := apitokens.Parse(test.text)
		if got, want := tok.Scheme, test.scheme; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := tok.Token(), test.name; got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		tok, err := apitokens.GetToken(ctx, apitokens.DefaultReaders, test.text)
		if err != nil {
			t.Errorf("failed to get token: %v", err)
		}

		if got, want := tok.Scheme, test.scheme; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := tok.Token(), test.value; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

}
