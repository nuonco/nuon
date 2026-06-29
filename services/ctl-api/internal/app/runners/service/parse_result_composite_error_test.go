package service

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func ptr(s string) *string { return &s }

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
		if got := parseResultCompositeError(req, job); got != nil {
			t.Fatalf("expected nil on success, got %+v", got)
		}
	})

	t.Run("no message yields nil", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{Success: false}
		if got := parseResultCompositeError(req, job); got != nil {
			t.Fatalf("expected nil when no message, got %+v", got)
		}
	})

	t.Run("unrecognised message yields nil", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr("some unrelated failure")},
		}
		if got := parseResultCompositeError(req, job); got != nil {
			t.Fatalf("expected nil on parse miss, got %+v", got)
		}
	})

	t.Run("aws permission error parsed with owner source and hint", func(t *testing.T) {
		req := &CreateRunnerJobExecutionResultRequest{
			Success:       false,
			ErrorMetadata: map[string]*string{"message": ptr(awsMsg)},
		}
		got := parseResultCompositeError(req, job)
		if got == nil {
			t.Fatal("expected a composite error, got nil")
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
}
