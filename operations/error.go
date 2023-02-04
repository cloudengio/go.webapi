// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"fmt"
	"net/http"
)

type Error struct {
	Err        error
	Status     string
	StatusCode int
	Attempts   int
}

func (err *Error) Error() string {
	if err.Err == nil {
		return err.Status
	}
	return fmt.Sprintf("%v: %v", err.Status, err.Err)
}

func handleError(err error, status string, statusCode int, attempts int) error {
	if err == nil && statusCode == http.StatusOK {
		return nil
	}
	return &Error{
		Err:        err,
		Status:     status,
		StatusCode: statusCode,
		Attempts:   attempts,
	}
}
