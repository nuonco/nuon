package controlplane

import (
	"context"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/errcapture"
)

func contextWithErrorOutput(output string) context.Context {
	capture := errcapture.New()
	zap.New(capture.Core()).Error(output)
	return errcapture.NewContext(context.Background(), capture)
}

func TestClientAdapterInjectsCapturedOutputOnFailure(t *testing.T) {
	client := &fakeClient{}
	adapter := &clientAdapter{Client: client}
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{ErrorMetadata: map[string]string{"message": "exit status 1"}}

	if _, err := adapter.CreateJobExecutionResult(contextWithErrorOutput("Error: AccessDenied"), "job", "execution", req); err != nil {
		t.Fatal(err)
	}
	if got := client.resultRequest.ErrorMetadata[errcapture.MetadataKey]; got != "Error: AccessDenied" {
		t.Fatalf("error_output = %q", got)
	}
	if got := client.resultRequest.ErrorMetadata["message"]; got != "exit status 1" {
		t.Fatalf("message = %q", got)
	}
}

func TestClientAdapterDoesNotOverwriteErrorOutput(t *testing.T) {
	client := &fakeClient{}
	adapter := &clientAdapter{Client: client}
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		ErrorMetadata: map[string]string{errcapture.MetadataKey: "handler output"},
	}

	if _, err := adapter.CreateJobExecutionResult(contextWithErrorOutput("captured output"), "job", "execution", req); err != nil {
		t.Fatal(err)
	}
	if got := client.resultRequest.ErrorMetadata[errcapture.MetadataKey]; got != "handler output" {
		t.Fatalf("error_output = %q", got)
	}
}

func TestClientAdapterLeavesSuccessfulResultUntouched(t *testing.T) {
	client := &fakeClient{}
	adapter := &clientAdapter{Client: client}
	req := &models.ServiceCreateRunnerJobExecutionResultRequest{Success: true}

	if _, err := adapter.CreateJobExecutionResult(contextWithErrorOutput("captured output"), "job", "execution", req); err != nil {
		t.Fatal(err)
	}
	if client.resultRequest.ErrorMetadata != nil {
		t.Fatalf("successful result metadata = %#v", client.resultRequest.ErrorMetadata)
	}
}
