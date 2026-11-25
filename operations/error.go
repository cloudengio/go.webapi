// Copyright 2023 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations

import (
	"fmt"
)

// Error is returned for a request that fails, either because of an error,
// Err, or because of an unsuccessful response, StatusCode.
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

// Unwrap returns Err.
func (err *Error) Unwrap() error {
	return err.Err
}

func handleError(err error, status string, statusCode int, isOK func(int) bool, attempts int) error {
	if err == nil && isOK(statusCode) {
		return nil
	}
	return &Error{
		Err:        err,
		Status:     status,
		StatusCode: statusCode,
		Attempts:   attempts,
	}
}
