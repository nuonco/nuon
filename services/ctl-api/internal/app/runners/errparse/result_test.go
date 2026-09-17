package errparse_test

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/all"
)

func TestParseRunnerJobResultProviderGate(t *testing.T) {
	runnerJob := &app.RunnerJob{
		Type:      app.RunnerJobTypeTerraformDeploy,
		OwnerType: "install_deploys",
		OwnerID:   "dpl123",
	}
	metadata := map[string]string{
		errparse.ErrorMetadataOutput: "Error: creating S3 Bucket: AccessDenied: User: arn:aws:iam::123:role/runner is not authorized to perform: s3:CreateBucket on resource: arn:aws:s3:::bucket",
	}

	tests := []struct {
		name     string
		provider errparse.Provider
		wantType string
	}{
		{name: "AWS parser applies to AWS jobs", provider: errparse.ProviderAWS, wantType: "terraform.aws_permission"},
		{name: "AWS parser does not apply to Azure jobs", provider: errparse.ProviderAzure, wantType: "terraform.error"},
		{name: "unknown provider fails open", provider: errparse.ProviderUnknown, wantType: "terraform.aws_permission"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := errparse.ParseRunnerJobResult(false, metadata, runnerJob, func() errparse.Provider {
				return test.provider
			})
			if err != nil {
				t.Fatalf("parse result: %v", err)
			}
			if got == nil || string(got.Type) != test.wantType {
				t.Fatalf("type = %v, want %q", got, test.wantType)
			}
		})
	}
}

func TestParseRunnerJobResultResolvesProviderLazily(t *testing.T) {
	resolved := false
	runnerJob := &app.RunnerJob{Type: app.RunnerJobTypeTerraformDeploy}
	got, err := errparse.ParseRunnerJobResult(false, map[string]string{
		errparse.ErrorMetadataMessage: "some unrelated failure",
	}, runnerJob, func() errparse.Provider {
		resolved = true
		return errparse.ProviderAWS
	})
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if got == nil || got.Type != "generic" {
		t.Fatalf("expected generic error, got %+v", got)
	}
	if resolved {
		t.Fatal("provider was resolved without a provider-specific signal")
	}
}

func TestParseRunnerJobResultActionToolGate(t *testing.T) {
	runnerJob := &app.RunnerJob{Type: app.RunnerJobTypeActionsWorkflowRun}

	got, err := errparse.ParseRunnerJobResult(false, map[string]string{
		errparse.ErrorMetadataOutput: "Error: ordinary action script failure",
	}, runnerJob, nil)
	if err != nil {
		t.Fatalf("parse action result: %v", err)
	}
	if got == nil || got.Type != "generic" {
		t.Fatalf("action error type = %v, want generic", got)
	}

	got, err = errparse.ParseRunnerJobResult(false, map[string]string{
		errparse.ErrorMetadataOutput: "AccessDenied: User: arn:aws:iam::123:role/runner is not authorized to perform: s3:CreateBucket on resource: arn:aws:s3:::bucket",
	}, runnerJob, func() errparse.Provider {
		return errparse.ProviderAWS
	})
	if err != nil {
		t.Fatalf("parse action provider result: %v", err)
	}
	if got == nil || got.Type != "terraform.aws_permission" {
		t.Fatalf("action provider error type = %v, want terraform.aws_permission", got)
	}
}
