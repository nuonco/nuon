package api

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/errcapture"
)

// fakeClient embeds the interface (nil) and records the request passed to
// CreateJobExecutionResult; only that method is exercised in these tests.
type fakeClient struct {
	nuonrunner.Client
	got *models.ServiceCreateRunnerJobExecutionResultRequest
}

func (f *fakeClient) CreateJobExecutionResult(ctx context.Context, jobID, jobExecutionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	f.got = req
	return &models.AppRunnerJobExecutionResult{}, nil
}

// ctxWithCapture returns a context carrying a Capture populated with output by
// logging through the real capture core (append is unexported).
func ctxWithCapture(output string) context.Context {
	c := errcapture.New()
	if output != "" {
		zap.New(c.Core()).Error(output)
	}
	return errcapture.NewContext(context.Background(), c)
}

func TestResultCapture_InjectsOnFailure(t *testing.T) {
	fake := &fakeClient{}
	c := &resultCaptureClient{Client: fake}

	ctx := ctxWithCapture("Error: AccessDenied s3:CreateBucket")
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:       false,
		ErrorMetadata: map[string]string{"message": "exit status 1"},
	}

	if _, err := c.CreateJobExecutionResult(ctx, "job", "exec", req); err != nil {
		t.Fatal(err)
	}
	if got := fake.got.ErrorMetadata[errcapture.MetadataKey]; got != "Error: AccessDenied s3:CreateBucket" {
		t.Fatalf("error_output = %q", got)
	}
	if fake.got.ErrorMetadata["message"] != "exit status 1" {
		t.Fatal("existing message must be preserved")
	}
}

// TestResultCapture_MultiLineTerraformDiagnostic mirrors the real failure that
// motivated this feature: terraform logs each AWS AccessDenied line separately
// at error level (via zaphclog), while the wrapped Go error is only a thin
// "exit status 1". The capture core must join the lines back into an
// error_output that still carries the full "is not authorized to perform"
// sentence the ctl-api AWS parser keys on.
func TestResultCapture_MultiLineTerraformDiagnostic(t *testing.T) {
	c := errcapture.New()
	l := zap.New(c.Core())
	for _, line := range []string{
		`Error: creating S3 Bucket (acme-nuon-clickhouse): operation error S3: CreateBucket, https response error StatusCode: 403, api error AccessDenied: User: arn:aws:sts::017500899169:assumed-role/broken-provision/component-deploy is not authorized to perform: s3:CreateBucket on resource: "arn:aws:s3:::acme-nuon-clickhouse" with an explicit deny in a permissions boundary: arn:aws:iam::017500899169:policy/broken-provision-boundary`,
		"terraform run errored",
	} {
		l.Error(line)
	}
	ctx := errcapture.NewContext(context.Background(), c)

	fake := &fakeClient{}
	cli := &resultCaptureClient{Client: fake}
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:       false,
		ErrorMetadata: map[string]string{"message": "exit status 1"},
	}
	if _, err := cli.CreateJobExecutionResult(ctx, "job", "exec", req); err != nil {
		t.Fatal(err)
	}

	out := fake.got.ErrorMetadata[errcapture.MetadataKey]
	for _, want := range []string{
		"is not authorized to perform: s3:CreateBucket",
		"AccessDenied",
		"terraform run errored",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("error_output missing %q; got %q", want, out)
		}
	}
}

func TestResultCapture_SuccessUntouched(t *testing.T) {
	fake := &fakeClient{}
	c := &resultCaptureClient{Client: fake}

	ctx := ctxWithCapture("some error output")
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{Success: true}

	if _, err := c.CreateJobExecutionResult(ctx, "job", "exec", req); err != nil {
		t.Fatal(err)
	}
	if _, ok := fake.got.ErrorMetadata[errcapture.MetadataKey]; ok {
		t.Fatal("successful results must not get error_output")
	}
}

func TestResultCapture_DoesNotOverwrite(t *testing.T) {
	fake := &fakeClient{}
	c := &resultCaptureClient{Client: fake}

	ctx := ctxWithCapture("captured output")
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:       false,
		ErrorMetadata: map[string]string{errcapture.MetadataKey: "handler-set output"},
	}

	if _, err := c.CreateJobExecutionResult(ctx, "job", "exec", req); err != nil {
		t.Fatal(err)
	}
	if got := fake.got.ErrorMetadata[errcapture.MetadataKey]; got != "handler-set output" {
		t.Fatalf("error_output overwritten: %q", got)
	}
}

func TestResultCapture_NoCaptureNoInjection(t *testing.T) {
	fake := &fakeClient{}
	c := &resultCaptureClient{Client: fake}

	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:       false,
		ErrorMetadata: map[string]string{"message": "exit status 1"},
	}

	if _, err := c.CreateJobExecutionResult(context.Background(), "job", "exec", req); err != nil {
		t.Fatal(err)
	}
	if _, ok := fake.got.ErrorMetadata[errcapture.MetadataKey]; ok {
		t.Fatal("no capture in ctx should mean no error_output")
	}
}

func TestResultCapture_NilMetadataInitialized(t *testing.T) {
	fake := &fakeClient{}
	c := &resultCaptureClient{Client: fake}

	ctx := ctxWithCapture("boom")
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{Success: false}

	if _, err := c.CreateJobExecutionResult(ctx, "job", "exec", req); err != nil {
		t.Fatal(err)
	}
	if got := fake.got.ErrorMetadata[errcapture.MetadataKey]; got != "boom" {
		t.Fatalf("error_output = %q; nil metadata should be initialized", got)
	}
}
