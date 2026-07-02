package service

import (
	"testing"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func ptr(s string) *string { return &s }

// fakeMetricsWriter records Incr calls; all other Writer methods are no-ops via
// the embedded (nil) interface, which is never called in these tests.
type fakeMetricsWriter struct {
	metrics.Writer
	incrs []incrCall
}

type incrCall struct {
	name string
	tags []string
}

func (f *fakeMetricsWriter) Incr(name string, tags []string) {
	f.incrs = append(f.incrs, incrCall{name: name, tags: tags})
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func TestParseResultCompositeError(t *testing.T) {
	job := &app.RunnerJob{OwnerType: "install_deploys", OwnerID: "idpl_123"}

	awsMsg := "Error: creating S3 Bucket: AccessDenied: User: " +
		"arn:aws:iam::123:role/nuon-runner is not authorized to perform: " +
		"s3:CreateBucket on resource: arn:aws:s3:::acme-prod-assets"

	t.Run("success yields nil", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{
			Success:       true,
			ErrorMetadata: map[string]*string{"message": ptr(awsMsg)},
		}
		if got := parseCompositeError(req, job); got != nil {
			t.Fatalf("expected nil on success, got %+v", got)
		}
	})

	t.Run("no message yields nil", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{Success: false}
		if got := parseCompositeError(req, job); got != nil {
			t.Fatalf("expected nil when no message, got %+v", got)
		}
	})

	t.Run("unrecognised message yields generic fallback", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr("some unrelated failure")},
		}
		got := parseCompositeError(req, job)
		if got == nil {
			t.Fatal("expected the generic fallback to produce a composite error")
		}
		if got.Type != "generic" {
			t.Fatalf("expected generic fallback type, got %q", got.Type)
		}
		if got.SourceType != "install_deploys" || got.SourceID != "idpl_123" {
			t.Fatalf("unexpected source: %s/%s", got.SourceType, got.SourceID)
		}
	})

	t.Run("aws permission error parsed with owner source and hint", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr(awsMsg)},
		}
		got := parseCompositeError(req, job)
		if got == nil {
			t.Fatal("expected a composite error, got nil")
		}
		if got.Type != "terraform.aws_permission" {
			t.Fatalf("expected AWS permission type to win over generic, got %q", got.Type)
		}
		if got.SourceType != "install_deploys" || got.SourceID != "idpl_123" {
			t.Fatalf("unexpected source: %s/%s", got.SourceType, got.SourceID)
		}
		if got.Version != compositeerrors.SchemaVersion {
			t.Fatalf("expected version %d, got %d", compositeerrors.SchemaVersion, got.Version)
		}
		if !got.Hints.SkipAutoRetry() {
			t.Fatal("expected skip_auto_retry hint to be captured")
		}
	})

	t.Run("prefers error_output over message", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{
			Success: false,
			ErrorMetadata: map[string]*string{
				"message":      ptr("exit status 1"),
				"error_output": ptr(awsMsg),
			},
		}
		got := parseCompositeError(req, job)
		if got == nil || got.Type != "terraform.aws_permission" {
			t.Fatalf("expected AWS permission parsed from error_output, got %+v", got)
		}
	})
}

func TestParseResultCompositeError_Metric(t *testing.T) {
	awsMsg := "Error: creating S3 Bucket: AccessDenied: User: " +
		"arn:aws:iam::123:role/nuon-runner is not authorized to perform: " +
		"s3:CreateBucket on resource: arn:aws:s3:::acme-prod-assets"

	tfDeploy := &app.RunnerJob{
		Type:      app.RunnerJobTypeTerraformDeploy,
		Group:     app.RunnerJobGroupDeploy,
		OwnerType: "install_deploys",
		OwnerID:   "idpl_123",
	}

	t.Run("specific match tags the matched type", func(t *testing.T) {
		mw := &fakeMetricsWriter{}
		s := &service{mw: mw}
		s.parseResultCompositeError(&CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr(awsMsg)},
		}, tfDeploy)

		if len(mw.incrs) != 1 {
			t.Fatalf("expected 1 metric, got %d", len(mw.incrs))
		}
		c := mw.incrs[0]
		if c.name != metricCompositeErrorParse {
			t.Fatalf("metric name = %q", c.name)
		}
		for _, want := range []string{"tool:terraform", "group:deploy", "matched_type:terraform.aws_permission"} {
			if !hasTag(c.tags, want) {
				t.Errorf("missing tag %q in %v", want, c.tags)
			}
		}
	})

	t.Run("parse miss is tagged generic", func(t *testing.T) {
		mw := &fakeMetricsWriter{}
		s := &service{mw: mw}
		s.parseResultCompositeError(&CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr("some unrelated failure")},
		}, tfDeploy)

		if len(mw.incrs) != 1 {
			t.Fatalf("expected 1 metric, got %d", len(mw.incrs))
		}
		if !hasTag(mw.incrs[0].tags, "matched_type:generic") {
			t.Errorf("expected matched_type:generic, got %v", mw.incrs[0].tags)
		}
	})

	t.Run("empty output is tagged miss", func(t *testing.T) {
		mw := &fakeMetricsWriter{}
		s := &service{mw: mw}
		s.parseResultCompositeError(&CreateRunnerJobExecutionResultRequest{Success: false}, tfDeploy)

		if len(mw.incrs) != 1 {
			t.Fatalf("expected 1 metric, got %d", len(mw.incrs))
		}
		if !hasTag(mw.incrs[0].tags, "matched_type:miss") {
			t.Errorf("expected matched_type:miss, got %v", mw.incrs[0].tags)
		}
	})

	t.Run("unknown tool falls back to unknown tag", func(t *testing.T) {
		mw := &fakeMetricsWriter{}
		s := &service{mw: mw}
		s.parseResultCompositeError(&CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr("boom")},
		}, &app.RunnerJob{OwnerType: "install_deploys", OwnerID: "x"})

		if !hasTag(mw.incrs[0].tags, "tool:unknown") || !hasTag(mw.incrs[0].tags, "group:unknown") {
			t.Errorf("expected unknown fallback tags, got %v", mw.incrs[0].tags)
		}
	})

	t.Run("success emits no metric", func(t *testing.T) {
		mw := &fakeMetricsWriter{}
		s := &service{mw: mw}
		s.parseResultCompositeError(&CreateRunnerJobExecutionResultRequest{Success: true}, tfDeploy)

		if len(mw.incrs) != 0 {
			t.Fatalf("expected no metric on success, got %d", len(mw.incrs))
		}
	})
}

func TestRunnerJobTool(t *testing.T) {
	cases := []struct {
		typ  app.RunnerJobType
		want errparse.Tool
	}{
		{app.RunnerJobTypeTerraformDeploy, errparse.ToolTerraform},
		{app.RunnerJobTypeTerraformModuleBuild, errparse.ToolTerraform},
		{app.RunnerJobTypeSandboxTerraform, errparse.ToolTerraform},
		{app.RunnerJobTypeSandboxTerraformPlan, errparse.ToolTerraform},
		{app.RunnerJobTypeRunnerTerraform, errparse.ToolTerraform},
		{app.RunnerJobTypeHelmChartDeploy, errparse.ToolHelm},
		{app.RunnerJobTypeHelmChartBuild, errparse.ToolHelm},
		{app.RunnerJobTypeRunnerHelm, errparse.ToolHelm},
		{app.RunnerJobTypePulumiDeploy, errparse.ToolPulumi},
		{app.RunnerJobTypePulumiBuild, errparse.ToolPulumi},
		{app.RunnerJobTypeSandboxPulumi, errparse.ToolPulumi},
		{app.RunnerJobTypeDockerBuild, errparse.ToolDocker},
		{app.RunnerJobTypeKubrenetesManifestDeploy, errparse.ToolKubernetes},
		{app.RunnerJobTypeKubernetesManifestBuild, errparse.ToolKubernetes},
		{app.RunnerJobTypeContainerImageBuild, errparse.ToolOCI},
		{app.RunnerJobTypeOCISync, errparse.ToolOCI},
		{app.RunnerJobTypeHealthCheck, errparse.ToolUnknown},
		{app.RunnerJobTypeNOOP, errparse.ToolUnknown},
		{app.RunnerJobTypeSandboxBuild, errparse.ToolUnknown},
	}
	for _, c := range cases {
		if got := runnerJobTool(&app.RunnerJob{Type: c.typ}); got != c.want {
			t.Errorf("runnerJobTool(%q) = %q, want %q", c.typ, got, c.want)
		}
	}
}

func TestRawErrorText(t *testing.T) {
	t.Run("prefers error_output over message", func(t *testing.T) {
		got := rawErrorText(map[string]*string{
			errMetaKeyMessage: ptr("exit status 1"),
			errMetaKeyOutput:  ptr("Error: creating S3 Bucket: AccessDenied"),
		})
		if got != "Error: creating S3 Bucket: AccessDenied" {
			t.Errorf("got %q, want the captured output", got)
		}
	})

	t.Run("falls back to message", func(t *testing.T) {
		got := rawErrorText(map[string]*string{errMetaKeyMessage: ptr("exit status 1")})
		if got != "exit status 1" {
			t.Errorf("got %q, want message fallback", got)
		}
	})

	t.Run("skips empty error_output", func(t *testing.T) {
		got := rawErrorText(map[string]*string{
			errMetaKeyOutput:  ptr(""),
			errMetaKeyMessage: ptr("exit status 1"),
		})
		if got != "exit status 1" {
			t.Errorf("got %q, want message when output is empty", got)
		}
	})

	t.Run("nil values and empty map", func(t *testing.T) {
		if got := rawErrorText(map[string]*string{errMetaKeyMessage: nil}); got != "" {
			t.Errorf("got %q, want empty for nil value", got)
		}
		if got := rawErrorText(nil); got != "" {
			t.Errorf("got %q, want empty for nil map", got)
		}
	})
}

func TestFlattenErrorMetadata(t *testing.T) {
	out := flattenErrorMetadata(map[string]*string{
		"message": ptr("boom"),
		"missing": nil,
		"step":    ptr("apply"),
	})
	if out["message"] != "boom" || out["step"] != "apply" {
		t.Errorf("unexpected map: %v", out)
	}
	if _, ok := out["missing"]; ok {
		t.Error("nil value should be dropped")
	}
	if flattenErrorMetadata(nil) != nil {
		t.Error("nil input should yield nil map")
	}
}
