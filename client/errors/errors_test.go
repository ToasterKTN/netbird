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
	if got := merr.Error(); got != "1 error occurred:\n\t* EOF" {
		t.Fatalf("unexpected format %q", got)
	}
	merr = Append(merr, Append(nil, os.ErrNotExist, os.ErrClosed))
	if len(merr.Errors) != 3 {
		t.Fatalf("nested errors must be flattened, got %d", len(merr.Errors))
	}
	if got := merr.Error(); got != "3 errors occurred:\n\t* EOF\n\t* file does not exist\n\t* file already closed" {
		t.Fatalf("unexpected format %q", got)
	}
	err := FormatErrorOrNil(merr)
	if !errors.Is(err, os.ErrNotExist) || !errors.Is(err, io.EOF) {
		t.Fatal("errors.Is must see wrapped errors")
	}
}
