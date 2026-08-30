package stackerrors

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestStackTemplateRenderError_implements(t *testing.T) {
	var _ compositeerrors.CompositeError = (*StackTemplateRenderError)(nil)
	var _ compositeerrors.HintsProvider = (*StackTemplateRenderError)(nil)
}

func TestStackTemplateRenderError_terminal(t *testing.T) {
	e := &StackTemplateRenderError{Platform: "aws", Detail: "some error"}
	if !e.Hints().Terminal() {
		t.Fatal("expected StackTemplateRenderError hints to be terminal")
	}
	if e.Hints().SkipAutoRetry() {
		t.Fatal("expected terminal hint, not skip_auto_retry")
	}
}

func TestStackTemplateRenderError_message(t *testing.T) {
	cases := []struct {
		e    *StackTemplateRenderError
		want string
	}{
		{&StackTemplateRenderError{Platform: "aws"}, "stack template rendering failed for aws"},
		{&StackTemplateRenderError{Platform: "azure"}, "stack template rendering failed for azure"},
		{&StackTemplateRenderError{}, "stack template rendering failed"},
	}
	for _, tc := range cases {
		if got := tc.e.Error(); got != tc.want {
			t.Errorf("Error() = %q, want %q", got, tc.want)
		}
	}
}

func TestStackTemplateRenderError_sections(t *testing.T) {
	e := &StackTemplateRenderError{Platform: "aws", Detail: "render boom"}
	sections := e.Sections()
	if len(sections) < 3 {
		t.Fatalf("expected at least 3 sections, got %d", len(sections))
	}

	found := false
	for _, s := range sections {
		if s.Kind == compositeerrors.SectionCode && s.Body == "render boom" {
			found = true
		}
	}
	if !found {
		t.Error("expected a code section with the error detail")
	}
}

func TestStackTemplateRenderError_noDetailSkipsCodeSection(t *testing.T) {
	e := &StackTemplateRenderError{Platform: "gcp"}
	for _, s := range e.Sections() {
		if s.Kind == compositeerrors.SectionCode {
			t.Errorf("unexpected code section when detail is empty: %+v", s)
		}
	}
}

func TestSandboxPlanRenderError_implements(t *testing.T) {
	var _ compositeerrors.CompositeError = (*SandboxPlanRenderError)(nil)
	var _ compositeerrors.HintsProvider = (*SandboxPlanRenderError)(nil)
}

func TestSandboxPlanRenderError_terminal(t *testing.T) {
	e := &SandboxPlanRenderError{Detail: "plan exploded"}
	if !e.Hints().Terminal() {
		t.Fatal("expected SandboxPlanRenderError hints to be terminal")
	}
}

func TestSandboxPlanRenderError_roundtrip(t *testing.T) {
	e := &SandboxPlanRenderError{Detail: "bad module path"}
	data, err := compositeerrors.New(e, compositeerrors.WithSource("install_sandbox_runs", "run123"))
	if err != nil {
		t.Fatalf("compositeerrors.New: %v", err)
	}
	if data.Type != SandboxPlanRenderErrorType {
		t.Errorf("type = %q, want %q", data.Type, SandboxPlanRenderErrorType)
	}
	if !data.Hints.Terminal() {
		t.Error("expected terminal hint after round-trip")
	}
	if data.SourceType != "install_sandbox_runs" || data.SourceID != "run123" {
		t.Errorf("source = %q/%q, want install_sandbox_runs/run123", data.SourceType, data.SourceID)
	}
}

func TestStackTemplateRenderError_roundtrip(t *testing.T) {
	e := &StackTemplateRenderError{Platform: "azure", Detail: "nested stack 404"}
	data, err := compositeerrors.New(e, compositeerrors.WithSource("install_stack_versions", "sv123"))
	if err != nil {
		t.Fatalf("compositeerrors.New: %v", err)
	}
	if data.Type != StackTemplateRenderErrorType {
		t.Errorf("type = %q, want %q", data.Type, StackTemplateRenderErrorType)
	}
	if !data.Hints.Terminal() {
		t.Error("expected terminal hint after round-trip")
	}
}
