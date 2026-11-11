package ui

import (
	"errors"
	"fmt"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
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

func TestClassifyErrorWrappedAPIUserError(t *testing.T) {
	description := "The install's labels match more than one app branch (acme-a, acme-b). Set app_branch on the install to choose one."
	err := fmt.Errorf("error resolving install create target: %w", &operations.GetInstallCreateTargetBadRequest{
		Payload: &models.StderrErrResponse{UserError: true, Error: "invalid request", Description: description},
	})

	code, msg := classifyError(err)
	if code != "user_error" {
		t.Fatalf("code = %q, want user_error", code)
	}
	if msg != description {
		t.Fatalf("msg = %q, want %q", msg, description)
	}
}
