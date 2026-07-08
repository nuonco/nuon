package helm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("unable to read fixture %s: %v", name, err)
	}
	return string(b)
}

func parse(raw string) compositeerrors.CompositeError {
	return errorParser{}.Parse(&errparse.ParseContext{Raw: raw})
}

func TestParse_StripsRunnerWrapper(t *testing.T) {
	ce := parse(readFixture(t, "reuse_name.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e, ok := ce.(*HelmError)
	if !ok {
		t.Fatalf("expected *HelmError, got %T", ce)
	}
	if e.Summary != "cannot reuse a name that is still in use" {
		t.Errorf("summary should be the SDK error with the runner wrapper stripped: %q", e.Summary)
	}
	if strings.Contains(e.Summary, "helm release") || strings.Contains(e.Summary, "unable to execute job") {
		t.Errorf("summary should not carry the runner wrapper nesting: %q", e.Summary)
	}
	if ce.Type() != HelmErrorType {
		t.Errorf("type = %q", ce.Type())
	}
	if ce.Severity() != compositeerrors.SeverityError {
		t.Errorf("severity = %q", ce.Severity())
	}
}

func TestParse_OwnershipConflict(t *testing.T) {
	ce := parse(readFixture(t, "ownership_conflict.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*HelmError)
	if !strings.HasPrefix(e.Summary, `ConfigMap "app-config"`) {
		t.Errorf("summary = %q", e.Summary)
	}
	if !strings.Contains(e.Summary, "exists and cannot be imported into the current release") {
		t.Errorf("summary missing the SDK cause: %q", e.Summary)
	}

	var output string
	for _, s := range ce.Sections() {
		if s.Heading == "Output" {
			output = s.Body
		}
	}
	if !strings.Contains(output, "invalid ownership metadata") {
		t.Errorf("Output missing helm detail: %q", output)
	}
}

// TestParse_WrapperWinsOverEarlierGenericLogLine is the key regression guard:
// the generic phrase "timed out waiting for the condition" appears in earlier
// streamed pod-log lines, but the headline must come from the runner's wrapper
// line, not the first line that merely mentions a generic phrase.
func TestParse_WrapperWinsOverEarlierGenericLogLine(t *testing.T) {
	ce := parse(readFixture(t, "wait_timeout_with_pod_logs.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*HelmError)
	if e.Summary != "timed out waiting for the condition" {
		t.Errorf("summary = %q", e.Summary)
	}
	if strings.Contains(e.Summary, "ImagePullBackOff") || strings.Contains(e.Summary, "pod api-server-0") {
		t.Errorf("summary should not be a streamed pod-log line: %q", e.Summary)
	}
	if !strings.Contains(e.Output, "ImagePullBackOff") {
		t.Errorf("Output should retain the streamed pod context: %q", e.Output)
	}
}

func TestParse_DryRunTemplateError(t *testing.T) {
	ce := parse(readFixture(t, "dry_run_template.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*HelmError)
	if !strings.HasPrefix(e.Summary, "template: mychart/templates/deployment.yaml") {
		t.Errorf("summary should strip the dry-run wrapper and lead with the template error: %q", e.Summary)
	}
}

// TestParse_CauseFallbackWithoutWrapper covers captured output that lost the
// runner wrapper but still carries a verified helm SDK cause string.
func TestParse_CauseFallbackWithoutWrapper(t *testing.T) {
	ce := parse("unable to build kubernetes objects from release manifest: error validating \"\": error validating data: apiVersion not set")
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e := ce.(*HelmError)
	if !strings.HasPrefix(e.Summary, "unable to build kubernetes objects from release manifest") {
		t.Errorf("summary = %q", e.Summary)
	}
}

func TestParse_NoHelmMarker(t *testing.T) {
	cases := []string{
		"",
		"exit status 1",
		"helm upgrade failed\nsome trailing context",
		"job step errored unable to execute job: unable to execute deploy pipeline",
		// A generic kubernetes phrase with no helm wrapper or cause marker must
		// defer to the generic parser, not produce a misleading helm error.
		"pod api-server-0: timed out waiting for the condition",
	}
	for _, in := range cases {
		if ce := parse(in); ce != nil {
			t.Errorf("parse(%q) = %v, want nil (no runner wrapper or verified helm cause)", in, ce)
		}
	}
}
