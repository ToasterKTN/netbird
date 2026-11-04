// Package errors provides a small multi-error type used to collect several
// failures (e.g. during cleanup) and report them together. It replaces
// github.com/hashicorp/go-multierror with the same Append/ErrorOrNil API.
package errors

import (
	"fmt"
	"strings"
)

// Error collects multiple errors. The zero value and nil are both empty.
type Error struct {
	Errors []error
}

// Append adds errs to err, flattening nested *Error values and skipping nil
// errors, and returns the (possibly newly allocated) *Error.
func Append(err *Error, errs ...error) *Error {
	if err == nil {
		err = &Error{}
	}
	for _, e := range errs {
		switch e := e.(type) {
		case nil:
		case *Error:
			if e != nil {
				err.Errors = append(err.Errors, e.Errors...)
			}
		default:
			err.Errors = append(err.Errors, e)
		}
	}
	return err
}

// Error implements the error interface.
func (e *Error) Error() string {
	return formatError(e.Errors)
}

// Unwrap returns the collected errors so errors.Is/As inspect each of them.
func (e *Error) Unwrap() []error {
	if e == nil {
		return nil
	}
	return e.Errors
}

// ErrorOrNil returns nil if no errors were collected, or e otherwise.
func (e *Error) ErrorOrNil() error {
	if e == nil || len(e.Errors) == 0 {
		return nil
	}
	return e
}

func formatError(es []error) string {
	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s", es[0])
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s",
		len(es), strings.Join(points, "\n\t"))
}

// FormatErrorOrNil returns nil if err holds no errors, or err otherwise.
func FormatErrorOrNil(err *Error) error {
	return err.ErrorOrNil()
}
