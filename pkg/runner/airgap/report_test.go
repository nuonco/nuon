package airgap

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestSuccessfulRunFinalizesStatusAndReport(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishWithArtifacts(t, client, models.AppRunnerJobGroupSandbox, "sandbox", true)
	finishWithArtifacts(t, client, models.AppRunnerJobGroupDeploy, "deploy-one", true)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-two", models.AppRunnerJobExecutionStatusFinished)

	select {
	case <-client.Done():
	case <-time.After(time.Second):
		t.Fatal("client did not complete")
	}

	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusFinished || status.FailedStep != "" || status.FinishedAt == nil {
		t.Fatalf("run status not finalized: %#v", status)
	}

	report := readReport(t, store)
	if report.Status != statestore.RunStatusFinished || len(report.Steps) != 3 {
		t.Fatalf("unexpected report: %#v", report)
	}
	byID := map[string]StepReport{}
	for _, step := range report.Steps {
		byID[step.ID] = step
	}
	if s := byID["sandbox"]; s.Success == nil || !*s.Success || s.Executions == 0 || string(s.Outputs) == "" {
		t.Fatalf("sandbox step missing result/outputs/executions: %#v", s)
	}
	if s := byID["deploy-two"]; s.Success == nil || !*s.Success {
		t.Fatalf("step without result.json should infer success from status: %#v", s)
	}
}

func TestFailedRunFinalizesStatusAndReport(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFailed)

	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusFailed || status.FailedStep != "sandbox" || status.FinishedAt == nil {
		t.Fatalf("failed run not finalized: %#v", status)
	}

	report := readReport(t, store)
	if report.Status != statestore.RunStatusFailed || report.FailedStep != "sandbox" {
		t.Fatalf("unexpected failed report: %#v", report)
	}
	for _, step := range report.Steps {
		if step.ID == "sandbox" && (step.Success == nil || *step.Success || step.Error == "") {
			t.Fatalf("failed step should report failure: %#v", step)
		}
	}
}

func TestResumeAfterFailureResetsRunStatus(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFailed)

	if _, err := NewClient(testEnvelope(), store, zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusInProgress || status.FailedStep != "" || status.FinishedAt != nil {
		t.Fatalf("resume should reset run status: %#v", status)
	}
}

// A run that failed before recording any outputs persists status.json with a
// null outputs map; the resumed client must still accept output writes.
func TestResumeAfterFailureWithoutOutputsAcceptsOutputs(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFailed)

	resumed, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishWithArtifacts(t, resumed, models.AppRunnerJobGroupSandbox, "sandbox", true)
	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := status.Outputs["sandbox"]; !ok {
		t.Fatalf("resumed run should record outputs: %#v", status.Outputs)
	}
}

func finishWithArtifacts(t *testing.T, client *Client, group models.AppRunnerJobGroup, id string, success bool) {
	t.Helper()
	ctx := context.Background()
	jobs, err := client.GetJobs(ctx, group, models.AppRunnerJobStatusAvailable, nil)
	if err != nil || len(jobs) != 1 || jobs[0].ID != id {
		t.Fatalf("unexpected next job: %#v %v", jobs, err)
	}
	execution, err := client.CreateJobExecution(ctx, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateJobExecutionResult(ctx, id, execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{Success: success}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateJobExecutionOutputs(ctx, id, execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{Outputs: map[string]any{"key": id}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateJobExecution(ctx, id, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished}); err != nil {
		t.Fatal(err)
	}
}

func readReport(t *testing.T, store statestore.Store) *Report {
	t.Helper()
	raw, ok, err := store.ReadReport()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("report.json was not written")
	}
	var report Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	return &report
}
