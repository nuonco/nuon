package airgap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func testEnvelope() *Envelope {
	return &Envelope{Version: "v0", InstallID: "install", Steps: []Step{
		{ID: "sandbox", Name: "sandbox", JobType: "sandbox-terraform", JobOperation: "exec", JobGroup: "sandbox", CompositePlan: []byte(`{}`)},
		{ID: "deploy-one", Name: "deploy one", JobType: "noop-deploy", JobOperation: "exec", JobGroup: "deploy", DependsOn: []string{"sandbox"}, CompositePlan: []byte(`{}`)},
		{ID: "deploy-two", Name: "deploy two", JobType: "noop-deploy", JobOperation: "exec", JobGroup: "deploy", DependsOn: []string{"deploy-one"}, CompositePlan: []byte(`{}`)},
	}}
}

func TestClientDependenciesAndCompletion(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	jobs, _ := client.GetJobs(ctx, models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 0 {
		t.Fatal("deploy should be dependency gated")
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFinished)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-one", models.AppRunnerJobExecutionStatusFinished)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-two", models.AppRunnerJobExecutionStatusFinished)
	select {
	case <-client.Done():
	case <-time.After(time.Second):
		t.Fatal("client did not complete")
	}
	if !client.Result().Succeeded {
		t.Fatal("run should succeed")
	}
}

func TestClientFailureHaltsDependents(t *testing.T) {
	client := newTestClient(t)
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFailed)
	jobs, _ := client.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 0 || client.Result().FailedStep != "sandbox" {
		t.Fatal("failure should halt dependent jobs")
	}
	select {
	case <-client.Done():
	default:
		t.Fatal("failure should complete run")
	}
}

func TestClientResumeSemantics(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFinished)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-one", models.AppRunnerJobExecutionStatusFailed)

	resumed, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	jobs, _ := resumed.GetJobs(ctx, models.AppRunnerJobGroupSandbox, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 0 {
		t.Fatal("finished sandbox step must not rerun on resume")
	}
	jobs, _ = resumed.GetJobs(ctx, models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 1 || jobs[0].ID != "deploy-one" {
		t.Fatalf("failed step should reset to available on resume, got %#v", jobs)
	}
	persisted, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range persisted.Steps {
		if st.ID == "deploy-one" && (st.Error != "" || st.ExecutionID != "" || st.StartedAt != nil || st.FinishedAt != nil) {
			t.Fatalf("reset step should clear execution metadata: %#v", st)
		}
	}
}

func TestClientResumeRejectsDifferentDeploymentID(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(testEnvelope(), store, zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	other := testEnvelope()
	other.InstallID = "inst-b"
	if _, err := NewClient(other, store, zap.NewNop()); err == nil {
		t.Fatal("resume with a different install ID must be rejected")
	}
}

func TestClientResumeAddsNewEnvelopeSteps(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	initial := testEnvelope()
	initial.Steps = initial.Steps[:1]
	client, err := NewClient(initial, store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFinished)

	resumed, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-resumed.Done():
		t.Fatal("new envelope steps must reopen a completed bootstrap")
	default:
	}
	jobs, err := resumed.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != "deploy-one" {
		t.Fatalf("new envelope step should be available after resume, got %#v", jobs)
	}
	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusInProgress || status.FinishedAt != nil {
		t.Fatalf("new envelope steps should reopen persisted status: %#v", status)
	}
}

func TestClientResumeReplansChainedPlanStep(t *testing.T) {
	envelope := func() *Envelope {
		return &Envelope{Version: "v0", InstallID: "install", Steps: []Step{
			{ID: "create", Name: "create", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", CompositePlan: []byte(`{"deploy_plan":{}}`)},
			{ID: "apply", Name: "apply", JobType: "terraform-deploy", JobOperation: "apply-plan", JobGroup: "deploy", DependsOn: []string{"create"}, PlanFromStep: "create", CompositePlan: []byte(`{"deploy_plan":{}}`)},
		}}
	}
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(envelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "create", models.AppRunnerJobExecutionStatusFinished)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "apply", models.AppRunnerJobExecutionStatusFailed)

	resumed, err := NewClient(envelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	jobs, _ := resumed.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 1 || jobs[0].ID != "create" {
		t.Fatalf("a failed apply must re-run its chained plan step first, got %#v", jobs)
	}
}

func TestClientResumeLateBindsPersistedResult(t *testing.T) {
	envelope := func() *Envelope {
		return &Envelope{Version: "v0", InstallID: "install", Steps: []Step{
			{ID: "create", Name: "create", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", CompositePlan: []byte(`{"sandbox_run_plan":{}}`)},
			{ID: "apply", Name: "apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", DependsOn: []string{"create"}, PlanFromStep: "create", CompositePlan: []byte(`{"sandbox_run_plan":{}}`)},
		}}
	}
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(envelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	execution, err := client.CreateJobExecution(ctx, "create", nil)
	if err != nil {
		t.Fatal(err)
	}
	planBytes := []byte("tfplan-raw-bytes")
	_, err = client.CreateJobExecutionResult(ctx, "create", execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:            true,
		ContentsCompressed: base64.URLEncoding.EncodeToString(planBytes),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.UpdateJobExecution(ctx, "create", execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatusFinished})
	if err != nil {
		t.Fatal(err)
	}

	resumed, err := NewClient(envelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := resumed.GetJobPlanJSON(ctx, "apply")
	if err != nil {
		t.Fatal(err)
	}
	var plan map[string]map[string]any
	if err := json.Unmarshal([]byte(rendered), &plan); err != nil {
		t.Fatal(err)
	}
	contents, _ := plan["sandbox_run_plan"]["apply_plan_contents"].(string)
	if contents != base64.StdEncoding.EncodeToString(planBytes) {
		t.Fatalf("resumed client did not late-bind persisted result, got %q", contents)
	}
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func finishNext(t *testing.T, client *Client, group models.AppRunnerJobGroup, id string, status models.AppRunnerJobExecutionStatus) {
	t.Helper()
	jobs, err := client.GetJobs(context.Background(), group, models.AppRunnerJobStatusAvailable, nil)
	if err != nil || len(jobs) != 1 || jobs[0].ID != id {
		t.Fatalf("unexpected next job: %#v %v", jobs, err)
	}
	execution, err := client.CreateJobExecution(context.Background(), id, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.UpdateJobExecution(context.Background(), id, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: status, StatusDescription: "failed"})
	if err != nil {
		t.Fatal(err)
	}
}
