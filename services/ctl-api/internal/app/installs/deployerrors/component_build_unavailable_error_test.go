package deployerrors

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestComponentBuildUnavailableError_Failed(t *testing.T) {
	e := &ComponentBuildUnavailableError{
		Reason:                 ComponentBuildUnavailableReasonFailed,
		ComponentID:            "cmp123",
		ComponentName:          "api",
		BuildID:                "bld456",
		BuildStatus:            "error",
		BuildStatusDescription: "build step failed",
	}

	if got, want := e.Type(), ComponentBuildUnavailableErrorType; got != want {
		t.Errorf("Type() = %q, want %q", got, want)
	}
	if e.Severity() != compositeerrors.SeverityError {
		t.Errorf("Severity() = %q, want error", e.Severity())
	}
	if got := e.Error(); !strings.Contains(got, "api") || !strings.Contains(got, "failed") {
		t.Errorf("Error() = %q, want it to mention the component and failure", got)
	}

	sections := e.Sections()
	if len(sections) < 2 {
		t.Fatalf("expected at least why + how-to-fix sections, got %d", len(sections))
	}

	var joined string
	for _, s := range sections {
		joined += s.Heading + "\n" + s.Body + "\n"
	}
	for _, want := range []string{"error", "build step failed", "bld456"} {
		if !strings.Contains(joined, want) {
			t.Errorf("sections missing %q; got:\n%s", want, joined)
		}
	}
	if !strings.Contains(joined, "How to fix") {
		t.Errorf("sections missing a how-to-fix heading; got:\n%s", joined)
	}
}

func TestComponentBuildUnavailableError_Missing(t *testing.T) {
	e := &ComponentBuildUnavailableError{
		Reason:        ComponentBuildUnavailableReasonMissing,
		ComponentName: "worker",
	}
	if got := e.Error(); !strings.Contains(got, "worker") {
		t.Errorf("Error() = %q, want it to mention the component", got)
	}

	data := compositeerrors.New(e)
	if data.Type != ComponentBuildUnavailableErrorType {
		t.Errorf("New().Type = %q, want %q", data.Type, ComponentBuildUnavailableErrorType)
	}
	if data.Message == "" {
		t.Error("New().Message should not be empty")
	}
}
