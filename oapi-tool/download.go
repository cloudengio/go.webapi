// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
)

type DownloadFlags struct {
	Output string `subcmd:"output,,'filename to save the downloaded file as, if not set, the contents will be written to stdout'"`
}

func downloadCmd(_ context.Context, values any, args []string) error {
	fv := values.(*DownloadFlags)
	buf := &bytes.Buffer{}
	downloadURL := args[0]
	if err := download(downloadURL, buf); err != nil {
		return err
	}
	return writeOutput(fv.Output, buf.Bytes())
}

func writeOutput(filename string, data []byte) error {
	if len(filename) > 0 {
		return os.WriteFile(filename, data, 0660)
	}
	_, err := os.Stdout.Write(data)
	return err
}

func download(downloadURL string, out io.Writer) error {
	if _, err := url.Parse(downloadURL); err != nil {
		return err
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, _ []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
