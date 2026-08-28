package customermanaged

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
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

func TestClientFailureAwaitsRetryDecision(t *testing.T) {
	client := newTestClient(t)
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFailed)
	jobs, _ := client.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 0 || client.Result().FailedStep != "sandbox" {
		t.Fatal("failure should halt dependent jobs")
	}
	status, err := client.store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusFailedPendingRetry || status.ResultDirective != statestore.DirectiveAwaitRetry {
		t.Fatalf("failure should await an explicit retry decision: %#v", status)
	}
	select {
	case <-client.Done():
		t.Fatal("failure should remain controllable")
	default:
	}
}

func TestClientRetryPreservesFailedAttempt(t *testing.T) {
	dir := t.TempDir()
	store, err := statestore.NewDisk(dir)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(testEnvelope(), store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox", models.AppRunnerJobExecutionStatusFailed)
	failedRunID := client.RunID()
	if err := client.ApplyControl(statestore.ControlActionRetry); err != nil {
		t.Fatal(err)
	}
	if client.RunID() == failedRunID {
		t.Fatal("retry reused the failed run ID")
	}
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(statestore.InstallRunStatusKey(failedRunID))))
	if err != nil {
		t.Fatal(err)
	}
	var failed statestore.Status
	if err := json.Unmarshal(raw, &failed); err != nil {
		t.Fatal(err)
	}
	if failed.Status != statestore.RunStatusFailed || failed.ResultDirective != statestore.DirectiveRetryGroup || failed.FinishedAt == nil {
		t.Fatalf("failed attempt was not finalized immutably: %#v", failed)
	}
	next, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if next.PreviousRunID != failedRunID || next.Status != statestore.RunStatusInProgress {
		t.Fatalf("retry attempt is not linked to failed attempt: %#v", next)
	}
}

func TestClientUserSkipSyncContinuesToPlan(t *testing.T) {
	client := newDirectiveTestClient(t)
	finishJob(t, client, "sync", models.AppRunnerJobExecutionStatusFailed)
	if err := client.ApplyControl(statestore.ControlActionUserSkip); err != nil {
		t.Fatal(err)
	}
	assertStepStatus(t, client, "sync", string(models.AppStatusUserDashSkipped), statestore.DirectiveContinue)
	assertStepStatus(t, client, "independent", string(models.AppRunnerJobStatusAvailable), "")
	jobs, _ := client.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 1 || jobs[0].ID != "plan" {
		t.Fatalf("sync skip should release its plan and independent work: %#v", jobs)
	}
}

func TestClientUserSkipPlanInvalidatesApplyDescendants(t *testing.T) {
	client := newDirectiveTestClient(t)
	finishJob(t, client, "sync", models.AppRunnerJobExecutionStatusFinished)
	finishJob(t, client, "plan", models.AppRunnerJobExecutionStatusFailed)
	if err := client.ApplyControl(statestore.ControlActionUserSkip); err != nil {
		t.Fatal(err)
	}
	assertStepStatus(t, client, "plan", string(models.AppStatusUserDashSkipped), statestore.DirectiveContinue)
	assertStepStatus(t, client, "apply", string(models.AppRunnerJobStatusNotDashAttempted), "")
	assertStepStatus(t, client, "after", string(models.AppRunnerJobStatusNotDashAttempted), "")
	jobs, _ := client.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 1 || jobs[0].ID != "independent" {
		t.Fatalf("plan skip should retain only independent work: %#v", jobs)
	}
}

func TestClientUserSkipApplyContinuesDependents(t *testing.T) {
	client := newDirectiveTestClient(t)
	finishJob(t, client, "sync", models.AppRunnerJobExecutionStatusFinished)
	finishJob(t, client, "plan", models.AppRunnerJobExecutionStatusFinished)
	finishJob(t, client, "apply", models.AppRunnerJobExecutionStatusFailed)
	if err := client.ApplyControl(statestore.ControlActionUserSkip); err != nil {
		t.Fatal(err)
	}
	assertStepStatus(t, client, "independent", string(models.AppRunnerJobStatusAvailable), "")
	jobs, _ := client.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if len(jobs) != 1 || jobs[0].ID != "after" {
		t.Fatalf("apply skip should release dependents without fabricating outputs: %#v", jobs)
	}
}

func TestClientCancelStopsAttempt(t *testing.T) {
	client := newDirectiveTestClient(t)
	finishJob(t, client, "sync", models.AppRunnerJobExecutionStatusFailed)
	if err := client.ApplyControl(statestore.ControlActionCancel); err != nil {
		t.Fatal(err)
	}
	status, err := client.store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusCancelled || status.ResultDirective != statestore.DirectiveStop || status.FinishedAt == nil {
		t.Fatalf("cancel did not terminalize attempt: %#v", status)
	}
	for _, step := range status.Steps {
		if step.ID != "sync" && step.Status != string(models.AppRunnerJobStatusNotDashAttempted) {
			t.Fatalf("cancel left %s as %s", step.ID, step.Status)
		}
	}
	if err := client.ApplyControl(statestore.ControlActionCancel); err == nil {
		t.Fatal("terminal attempt accepted another cancel")
	}
}

func newDirectiveTestClient(t *testing.T) *Client {
	t.Helper()
	envelope := &Envelope{Version: "v0", InstallID: "install", Steps: []Step{
		{ID: "sync", Name: "sync", JobType: "sync-oci", JobOperation: "sync", JobGroup: "sync", CompositePlan: []byte(`{}`)},
		{ID: "plan", Name: "plan", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"sync"}, CompositePlan: []byte(`{}`)},
		{ID: "apply", Name: "apply", JobType: "terraform-deploy", JobOperation: "apply-plan", JobGroup: "deploy", DependsOn: []string{"plan"}, PlanFromStep: "plan", CompositePlan: []byte(`{}`)},
		{ID: "after", Name: "after", JobType: "noop-deploy", JobOperation: "exec", JobGroup: "deploy", DependsOn: []string{"apply"}, CompositePlan: []byte(`{}`)},
		{ID: "independent", Name: "independent", JobType: "noop-deploy", JobOperation: "exec", JobGroup: "deploy", CompositePlan: []byte(`{}`)},
	}}
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(envelope, store, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func assertStepStatus(t *testing.T, client *Client, id, status string, directive statestore.ResultDirective) {
	t.Helper()
	persisted, err := client.store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	step := findStepStatus(persisted.Steps, id)
	if step == nil || step.Status != status || step.ResultDirective != directive {
		t.Fatalf("unexpected %s status: %#v", id, step)
	}
}

func finishJob(t *testing.T, client *Client, id string, status models.AppRunnerJobExecutionStatus) {
	t.Helper()
	execution, err := client.CreateJobExecution(context.Background(), id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateJobExecution(context.Background(), id, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: status, StatusDescription: "failed"}); err != nil {
		t.Fatal(err)
	}
}

func TestClientBundleUpgradePlansChangedComponentBeforeApproval(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	envelope := &Envelope{Version: "v0", InstallID: "install", Steps: []Step{
		{ID: "sandbox-apply", Name: "sandbox apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox"},
		{ID: "sync-api", Name: "api sync", JobType: "oci-sync", JobOperation: "exec", JobGroup: "sync", DependsOn: []string{"sandbox-apply"}},
		{ID: "deploy-api-plan", Name: "api plan", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"sync-api"}},
		{ID: "deploy-api-apply", Name: "api apply", JobType: "terraform-deploy", JobOperation: "apply-plan", JobGroup: "deploy", DependsOn: []string{"deploy-api-plan"}},
		{ID: "sync-worker", Name: "worker sync", JobType: "oci-sync", JobOperation: "exec", JobGroup: "sync", DependsOn: []string{"deploy-api-apply"}},
	}}
	now := time.Now().UTC()
	if err := store.WriteStatus(&statestore.Status{
		InstallID: "install", BundleDigest: "sha256:v1", RunID: "v1", Status: statestore.RunStatusFinished,
		StartedAt: now.Add(-time.Hour), FinishedAt: &now,
		Steps: []statestore.StepStatus{
			{ID: "sandbox-apply", Status: string(models.AppRunnerJobStatusFinished)},
			{ID: "sync-api", Status: string(models.AppRunnerJobStatusFinished)},
			{ID: "deploy-api-plan", Status: string(models.AppRunnerJobStatusFinished)},
			{ID: "deploy-api-apply", Status: string(models.AppRunnerJobStatusFinished)},
			{ID: "sync-worker", Status: string(models.AppRunnerJobStatusFinished)},
		},
	}); err != nil {
		t.Fatal(err)
	}

	client, err := NewClientWithOptions(envelope, store, zap.NewNop(), ClientOptions{
		BundleDigest: "sha256:v2", BundleUpgrade: true, UpgradeComponents: []string{"api"}, RequireApproval: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	upgradeStatus, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if upgradeStatus.RunType != statestore.RunTypeUpgrade || upgradeStatus.PreviousRunID != "v1" {
		t.Fatalf("upgrade history linkage was not persisted: %#v", upgradeStatus)
	}
	var workerSkipped bool
	for _, step := range upgradeStatus.Steps {
		if step.ID == "sync-worker" && step.Status == string(models.AppStatusAutoDashSkipped) && step.SourceRunID == "v1" {
			workerSkipped = true
		}
	}
	if !workerSkipped {
		t.Fatal("unchanged worker step was not auto-skipped with successful source provenance")
	}
	finishNext(t, client, models.AppRunnerJobGroupSync, "sync-api", models.AppRunnerJobExecutionStatusFinished)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-api-plan", models.AppRunnerJobExecutionStatusFinished)
	jobs, err := client.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if err != nil || len(jobs) != 0 {
		t.Fatalf("apply should wait for approval: %#v %v", jobs, err)
	}
	if err := client.ApproveApply(); err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-api-apply", models.AppRunnerJobExecutionStatusFinished)
	select {
	case <-client.Done():
	case <-time.After(time.Second):
		t.Fatal("upgrade did not complete")
	}
	restarted, err := NewClientWithOptions(envelope, store, zap.NewNop(), ClientOptions{BundleDigest: "sha256:v2"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-restarted.Done():
	default:
		t.Fatal("completed upgrade should remain complete after restart")
	}
	jobs, err = restarted.GetJobs(context.Background(), models.AppRunnerJobGroupDeploy, models.AppRunnerJobStatusAvailable, nil)
	if err != nil || len(jobs) != 0 {
		t.Fatalf("completed upgrade should not replay apply: %#v %v", jobs, err)
	}
}

func TestClientBundleUpgradeDoesNotSkipUnsuccessfulPriorStep(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.WriteStatus(&statestore.Status{
		InstallID: "install", BundleDigest: "sha256:v1", RunID: "failed-run", Status: statestore.RunStatusFailed,
		StartedAt: now.Add(-time.Hour), FinishedAt: &now,
		Steps: []statestore.StepStatus{{ID: "sandbox", Status: string(models.AppRunnerJobStatusFailed)}},
	}); err != nil {
		t.Fatal(err)
	}

	_, err = NewClientWithOptions(testEnvelope(), store, zap.NewNop(), ClientOptions{
		BundleDigest: "sha256:v2", BundleUpgrade: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range status.Steps {
		if step.Status != string(models.AppRunnerJobStatusAvailable) || step.SourceRunID != "" {
			t.Fatalf("step %q inherited an unsuccessful prior execution: %#v", step.ID, step)
		}
	}
}

func TestClientBundleUpgradeWithoutComponentChangesStillRequiresApproval(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.WriteStatus(&statestore.Status{
		InstallID: "install", BundleDigest: "sha256:v1", RunID: "v1", Status: statestore.RunStatusFinished,
		StartedAt: now.Add(-time.Hour), FinishedAt: &now,
		Steps: []statestore.StepStatus{
			{ID: "sandbox", Name: "sandbox", Status: string(models.AppRunnerJobStatusFinished)},
			{ID: "deploy-one", Name: "deploy one", Status: string(models.AppRunnerJobStatusFinished)},
			{ID: "deploy-two", Name: "deploy two", Status: string(models.AppRunnerJobStatusFinished)},
		},
	}); err != nil {
		t.Fatal(err)
	}
	client, err := NewClientWithOptions(testEnvelope(), store, zap.NewNop(), ClientOptions{BundleDigest: "sha256:v2", BundleUpgrade: true, RequireApproval: true})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-client.Done():
		t.Fatal("bundle upgrade completed without approval")
	default:
	}
	if err := client.ApproveApply(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-client.Done():
	case <-time.After(time.Second):
		t.Fatal("approved bundle upgrade did not complete")
	}
}

func TestClientBundleUpgradeUsesSeparateSandboxAndComponentApprovals(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := store.WriteStatus(&statestore.Status{InstallID: "install", BundleDigest: "sha256:v1", RunID: "v1", Status: statestore.RunStatusFinished, StartedAt: now}); err != nil {
		t.Fatal(err)
	}
	envelope := &Envelope{Version: "v0", InstallID: "install", Steps: []Step{
		{ID: "sandbox-plan", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox"},
		{ID: "sandbox-apply", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", DependsOn: []string{"sandbox-plan"}},
		{ID: "sync-api", JobType: "oci-sync", JobOperation: "exec", JobGroup: "sync", DependsOn: []string{"sandbox-apply"}},
		{ID: "deploy-api-plan", JobType: "terraform-deploy", JobOperation: "create-apply-plan", JobGroup: "deploy", DependsOn: []string{"sync-api"}},
		{ID: "deploy-api-apply", JobType: "terraform-deploy", JobOperation: "apply-plan", JobGroup: "deploy", DependsOn: []string{"deploy-api-plan"}},
	}}
	client, err := NewClientWithOptions(envelope, store, zap.NewNop(), ClientOptions{
		BundleDigest: "sha256:v2", BundleUpgrade: true, UpgradeSandbox: true,
		UpgradeComponents: []string{"api"}, RequireApproval: true, ApprovalPhase: "sandbox",
	})
	if err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox-plan", models.AppRunnerJobExecutionStatusFinished)
	if err := client.ApproveApply(); err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupSandbox, "sandbox-apply", models.AppRunnerJobExecutionStatusFinished)
	if required, phase := client.Approval(); !required || phase != "components" {
		t.Fatalf("component approval should follow sandbox apply, got required=%t phase=%q", required, phase)
	}
	finishNext(t, client, models.AppRunnerJobGroupSync, "sync-api", models.AppRunnerJobExecutionStatusFinished)
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-api-plan", models.AppRunnerJobExecutionStatusFinished)
	if err := client.ApproveApply(); err != nil {
		t.Fatal(err)
	}
	finishNext(t, client, models.AppRunnerJobGroupDeploy, "deploy-api-apply", models.AppRunnerJobExecutionStatusFinished)
	select {
	case <-client.Done():
	case <-time.After(time.Second):
		t.Fatal("phased bundle upgrade did not complete")
	}
}

func TestClientRejectsChangedBundleWithoutCandidate(t *testing.T) {
	store, err := statestore.NewDisk(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.WriteStatus(&statestore.Status{InstallID: "install", BundleDigest: "sha256:v1", RunID: "v1", Status: statestore.RunStatusFinished}); err != nil {
		t.Fatal(err)
	}
	_, err = NewClientWithOptions(testEnvelope(), store, zap.NewNop(), ClientOptions{BundleDigest: "sha256:v2"})
	if err == nil {
		t.Fatal("changed bundle without a staged candidate must be rejected")
	}
}

func TestClientRetryCreatesNewAttempt(t *testing.T) {
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
	previousRunID := resumed.RunID()
	if err := resumed.ApplyControl(statestore.ControlActionRetry); err != nil {
		t.Fatal(err)
	}
	if resumed.RunID() == previousRunID {
		t.Fatal("retry must create a new attempt")
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

func TestClientCompletedAttemptIsNotReopenedByEnvelopeDrift(t *testing.T) {
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
	default:
		t.Fatal("completed attempt must remain terminal")
	}
	status, err := store.ReadStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != statestore.RunStatusFinished || len(status.Steps) != 1 {
		t.Fatalf("completed attempt was rewritten by envelope drift: %#v", status)
	}
}

func TestClientRetryReplansChainedPlanStep(t *testing.T) {
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
	if err := resumed.ApplyControl(statestore.ControlActionRetry); err != nil {
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
