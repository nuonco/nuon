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

func helmErr(t *testing.T, ce compositeerrors.CompositeError) *HelmError {
	t.Helper()
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e, ok := ce.(*HelmError)
	if !ok {
		t.Fatalf("expected *HelmError, got %T", ce)
	}
	return e
}

func TestParse_StripsRunnerWrapperAndClassifies(t *testing.T) {
	ce := parse(readFixture(t, "reuse_name.txt"))
	e := helmErr(t, ce)
	if e.Summary != "cannot reuse a name that is still in use" {
		t.Errorf("summary should be the SDK error with the runner wrapper stripped: %q", e.Summary)
	}
	if strings.Contains(e.Summary, "helm release") || strings.Contains(e.Summary, "unable to execute job") {
		t.Errorf("summary should not carry the runner wrapper nesting: %q", e.Summary)
	}
	if ce.Type() != HelmNameInUseType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmNameInUseType)
	}
	if e.Reason != "name_in_use" {
		t.Errorf("reason = %q", e.Reason)
	}
	if ce.Severity() != compositeerrors.SeverityError {
		t.Errorf("severity = %q", ce.Severity())
	}
	if !e.Hints().SkipAutoRetry() {
		t.Error("a name-in-use failure should set skip_auto_retry")
	}
}

func TestParse_ImmutableField(t *testing.T) {
	ce := parse(readFixture(t, "immutable_field.txt"))
	e := helmErr(t, ce)
	if ce.Type() != HelmImmutableFieldType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmImmutableFieldType)
	}
	if !strings.HasPrefix(e.Summary, `cannot patch "clickhouse"`) {
		t.Errorf("summary = %q", e.Summary)
	}
	if !e.Hints().SkipAutoRetry() {
		t.Error("an immutable-field failure should set skip_auto_retry")
	}
}

func TestParse_OwnershipConflict(t *testing.T) {
	ce := parse(readFixture(t, "ownership_conflict.txt"))
	e := helmErr(t, ce)
	if ce.Type() != HelmOwnershipConflictType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmOwnershipConflictType)
	}
	if !strings.HasPrefix(e.Summary, `ConfigMap "app-config"`) {
		t.Errorf("summary = %q", e.Summary)
	}
	if !e.Hints().SkipAutoRetry() {
		t.Error("an ownership conflict should set skip_auto_retry")
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

func TestParse_HookFailedIsRetryable(t *testing.T) {
	ce := parse(readFixture(t, "hook_failed.txt"))
	e := helmErr(t, ce)
	if ce.Type() != HelmHookFailedType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmHookFailedType)
	}
	if e.Hints().SkipAutoRetry() {
		t.Error("a hook failure can be transient and should stay auto-retryable")
	}
}

// TestParse_WrapperWinsOverEarlierGenericLogLine is the key regression guard:
// the generic phrase "timed out waiting for the condition" appears in earlier
// streamed pod-log lines, but the headline must come from the runner's wrapper
// line, not the first line that merely mentions a generic phrase.
func TestParse_WrapperWinsOverEarlierGenericLogLine(t *testing.T) {
	ce := parse(readFixture(t, "wait_timeout_with_pod_logs.txt"))
	e := helmErr(t, ce)
	if ce.Type() != HelmWaitTimeoutType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmWaitTimeoutType)
	}
	if e.Summary != "timed out waiting for the condition" {
		t.Errorf("summary = %q", e.Summary)
	}
	if strings.Contains(e.Summary, "ImagePullBackOff") || strings.Contains(e.Summary, "pod api-server-0") {
		t.Errorf("summary should not be a streamed pod-log line: %q", e.Summary)
	}
	if !strings.Contains(e.Output, "ImagePullBackOff") {
		t.Errorf("Output should retain the streamed pod context: %q", e.Output)
	}
	if e.Hints().SkipAutoRetry() {
		t.Error("a wait timeout is often transient and should stay auto-retryable")
	}
}

func TestParse_DryRunTemplateError(t *testing.T) {
	ce := parse(readFixture(t, "dry_run_template.txt"))
	e := helmErr(t, ce)
	if ce.Type() != HelmRenderErrorType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmRenderErrorType)
	}
	if !strings.HasPrefix(e.Summary, "template: mychart/templates/deployment.yaml") {
		t.Errorf("summary should strip the dry-run wrapper and lead with the template error: %q", e.Summary)
	}
	if !e.Hints().SkipAutoRetry() {
		t.Error("a render error is deterministic and should set skip_auto_retry")
	}
}

// TestParse_CauseFallbackWithoutWrapper covers captured output that lost the
// runner wrapper but still carries a verified helm SDK cause string.
func TestParse_CauseFallbackWithoutWrapper(t *testing.T) {
	ce := parse("unable to build kubernetes objects from release manifest: error validating \"\": error validating data: apiVersion not set")
	e := helmErr(t, ce)
	if ce.Type() != HelmRenderErrorType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmRenderErrorType)
	}
	if !strings.HasPrefix(e.Summary, "unable to build kubernetes objects from release manifest") {
		t.Errorf("summary = %q", e.Summary)
	}
}

func TestParse_UnclassifiedHelmFailureFallsBackToGeneric(t *testing.T) {
	ce := parse("unable to upgrade helm release: some novel helm failure we have not classified")
	e := helmErr(t, ce)
	if ce.Type() != HelmErrorType {
		t.Errorf("type = %q, want the generic %q", ce.Type(), HelmErrorType)
	}
	if e.Reason != "" {
		t.Errorf("generic fallback should have no reason, got %q", e.Reason)
	}
	if e.Hints().SkipAutoRetry() {
		t.Error("generic helm failure should not force skip_auto_retry")
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
