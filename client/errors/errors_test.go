package errors

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestAppendAndFormat(t *testing.T) {
	var merr *Error
	if FormatErrorOrNil(merr) != nil {
		t.Fatal("nil *Error must yield nil")
	}
	merr = Append(merr, nil)
	if merr.ErrorOrNil() != nil {
		t.Fatal("nil errors must be skipped")
	}
	merr = Append(merr, io.EOF)
	if got := merr.Error(); got != "1 error occurred:\n\t* EOF\n\n" {
		t.Fatalf("unexpected format %q", got)
	}
	merr = Append(merr, Append(nil, os.ErrNotExist, os.ErrClosed))
	if len(merr.Errors) != 3 {
		t.Fatalf("nested errors must be flattened, got %d", len(merr.Errors))
	}
	if got := merr.Error(); got != "3 errors occurred:\n\t* EOF\n\t* file does not exist\n\t* file already closed\n\n" {
		t.Fatalf("unexpected format %q", got)
	}
	err := FormatErrorOrNil(merr)
	if !errors.Is(err, os.ErrNotExist) || !errors.Is(err, io.EOF) {
		t.Fatal("errors.Is must see wrapped errors")
	}
}

// Expected strings are go-multierror v1.1.1 output for the same inputs.
func TestFormatErrorOrNilMatchesUpstream(t *testing.T) {
	var merr *Error
	merr = Append(merr, io.EOF, os.ErrNotExist)
	err := FormatErrorOrNil(merr)
	if got := err.Error(); got != "2 errors occurred:\n\t* EOF\n\t* file does not exist" {
		t.Fatalf("compact format: %q", got)
	}
	if FormatErrorOrNil(nil) != nil || FormatErrorOrNil(&Error{}) != nil {
		t.Fatal("empty must be nil")
	}
}
