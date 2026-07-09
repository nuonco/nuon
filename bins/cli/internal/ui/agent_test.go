package ui

import (
	"errors"
	"fmt"
	"testing"
)

func TestClassifyErrorExitCode(t *testing.T) {
	err := &ErrExitCode{
		Err:  errors.New("2 of 3 component build(s) did not complete"),
		Code: "builds_failed",
		Exit: 3,
	}

	code, msg := classifyError(err)
	if code != "builds_failed" {
		t.Fatalf("code = %q, want builds_failed", code)
	}
	if msg != err.Error() {
		t.Fatalf("msg = %q, want %q", msg, err.Error())
	}

	wrapped := fmt.Errorf("context: %w", err)
	code, _ = classifyError(wrapped)
	if code != "builds_failed" {
		t.Fatalf("wrapped code = %q, want builds_failed", code)
	}

	code, _ = classifyError(errors.New("boom"))
	if code != "error" {
		t.Fatalf("plain error code = %q, want error", code)
	}
}
