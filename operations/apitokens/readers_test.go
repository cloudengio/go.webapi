// Copyright 2024 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package apitokens_test

/*
func TestTokenReaders(t *testing.T) {
	ctx := context.Background()
	os.Setenv("something-unlikely", "env-secret")
	for _, test := range []struct {
		text   string
		scheme string
		name   string
		value  []byte
	}{
		{"file://testdata/example.tok", "file", "testdata/example.tok", []byte("example-token\n")},
		{"env://something-unlikely", "env", "something-unlikely", []byte("env-secret")},
		{"literal://right-here", "literal", "right-here", []byte("right-here")},
	} {
		tok := apitokens.New(test.text)
		if got, want := tok.Scheme, test.scheme; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := tok.Path, test.name; got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		err := tok.Read(ctx, apitokens.DefaultReaders)
		if err != nil {
			t.Errorf("failed to get token: %v", err)
		}

		if got, want := tok.Scheme, test.scheme; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := tok.Token(), test.value; !bytes.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

}
*/
