package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
)

func TestEmitJSONErrorRendersConfigError(t *testing.T) {
	cfgErr := config.ErrConfig{
		Description: `branch "seed": one install group must be default`,
		Err:         errors.New(`branch "seed": one install group must be default`),
	}

	out := captureStdout(t, func() {
		_ = emitJSONError(fmt.Errorf("unable to load branch configs: %w", cfgErr))
	})

	var got jsonError
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", out, err)
	}
	want := fmt.Sprintf("%s %s", cfgErr.Description, cfgErr.Error())
	if got.Error != want {
		t.Fatalf("error = %q, want %q", got.Error, want)
	}
}

func TestEmitJSONErrorRendersCLIUserError(t *testing.T) {
	out := captureStdout(t, func() {
		_ = emitJSONError(&CLIUserError{Msg: "--file is required"})
	})

	var got jsonError
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", out, err)
	}
	if got.Error != "--file is required" {
		t.Fatalf("error = %q, want --file is required", got.Error)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(out)
}
