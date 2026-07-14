package controlplane

import (
	"context"
	"errors"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestTerminalExecutionError(t *testing.T) {
	if err := terminalExecutionError(models.AppRunnerJobExecutionStatusFinished); err != nil {
		t.Fatalf("finished status returned error: %v", err)
	}
	if err := terminalExecutionError(models.AppRunnerJobExecutionStatusCancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled status error = %v", err)
	}
	if err := terminalExecutionError(models.AppRunnerJobExecutionStatusTimedDashOut); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timed-out status error = %v", err)
	}
}

type fakeClient struct {
	planJSON      string
	compositePlan *models.PlantypesCompositePlan

	planCalls          int
	compositePlanCalls int
	outputs            any
	resultCalls        int
	resultRequest      *models.ServiceCreateRunnerJobExecutionResultRequest
	statuses           []models.AppRunnerJobExecutionStatus
	outputsCalls       int
}

func (c *fakeClient) GetJob(context.Context, string) (*models.AppRunnerJob, error) {
	return &models.AppRunnerJob{Status: models.AppRunnerJobStatusInDashProgress}, nil
}

func (c *fakeClient) GetJobPlanJSON(ctx context.Context, jobID string) (string, error) {
	c.planCalls++
	return c.planJSON, nil
}

func (c *fakeClient) GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error) {
	c.compositePlanCalls++
	return c.compositePlan, nil
}

func (c *fakeClient) UpdateJobExecution(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
	c.statuses = append(c.statuses, req.Status)
	return &models.AppRunnerJobExecution{}, nil
}

func (c *fakeClient) CreateJobExecutionResult(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	c.resultCalls++
	c.resultRequest = req
	return &models.AppRunnerJobExecutionResult{}, nil
}

func (c *fakeClient) CreateJobExecutionOutputs(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error) {
	c.outputsCalls++
	c.outputs = req.Outputs
	return &models.AppRunnerJobExecutionOutputs{}, nil
}

func (c *fakeClient) WriteControlPlaneLogs(ctx context.Context, logStreamID string, records []OTELLogRecord) error {
	return nil
}

func (c *fakeClient) WriteControlPlaneTraces(ctx context.Context, runnerID string, records []OTELTraceRecord) error {
	return nil
}

func TestExecuteSandboxBuildShortCircuits(t *testing.T) {
	client := &fakeClient{
		planJSON: `{"sandbox_mode":{"enabled":true,"outputs":{"image":{"tag":"v1.2.3","repository":"nuon/app-service"}}}}`,
	}
	e, err := NewExecutor(client, nil, Config{})
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}

	job := &models.AppRunnerJob{ID: "job1", Type: models.AppRunnerJobTypeContainerDashImageDashBuild}
	execution := &models.AppRunnerJobExecution{ID: "exec1"}

	if err := e.Execute(context.Background(), job, execution); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if client.planCalls != 1 {
		t.Errorf("expected plan fetched exactly once, got %d", client.planCalls)
	}
	if client.outputsCalls != 1 {
		t.Errorf("expected outputs written once, got %d", client.outputsCalls)
	}
	if client.resultCalls != 1 {
		t.Errorf("expected result written once, got %d", client.resultCalls)
	}
	outputs, ok := client.outputs.(map[string]any)
	if !ok {
		t.Fatalf("expected map outputs, got %#v", client.outputs)
	}
	img, ok := outputs["image"].(map[string]any)
	if !ok || img["tag"] != "v1.2.3" {
		t.Errorf("expected sandbox outputs propagated, got %#v", client.outputs)
	}
	if len(client.statuses) == 0 || client.statuses[len(client.statuses)-1] != models.AppRunnerJobExecutionStatusFinished {
		t.Errorf("expected execution to finish, got statuses %v", client.statuses)
	}
}

func TestExecuteNonSandboxDoesNotShortCircuit(t *testing.T) {
	client := &fakeClient{
		planJSON:      `{"sandbox_mode":{"enabled":false}}`,
		compositePlan: &models.PlantypesCompositePlan{BuildPlan: &models.PlantypesBuildPlan{}},
	}
	e, err := NewExecutor(client, nil, Config{})
	if err != nil {
		t.Fatalf("NewExecutor: %v", err)
	}

	job := &models.AppRunnerJob{ID: "job1", Type: models.AppRunnerJobTypeContainerDashImageDashBuild}
	execution := &models.AppRunnerJobExecution{ID: "exec1"}

	// A non-sandbox container-image build will fail on the real fetch path
	// (no reachable plan/registry here); the point is only that it does NOT
	// short-circuit as a sandbox build (no outputs/result written).
	_ = e.Execute(context.Background(), job, execution)

	if client.compositePlanCalls != 1 {
		t.Errorf("expected composite plan fetched exactly once, got %d", client.compositePlanCalls)
	}
	if client.outputsCalls != 0 {
		t.Errorf("non-sandbox build should not write sandbox outputs, got %d", client.outputsCalls)
	}
}
