package policy

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/policyerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type policySignal struct{}

func (*policySignal) Type() signal.SignalType         { return "test-policy" }
func (*policySignal) Validate(workflow.Context) error { return nil }
func (*policySignal) Execute(workflow.Context) error  { return nil }
func (*policySignal) RequiresPolicyEvaluation() bool  { return true }
func (*policySignal) AutoApproveOnPoliciesPassing(workflow.Context) bool {
	return true
}

// autoApproveOnlySignal has no policy target to evaluate — it only declares the
// auto-approve capability, like an install group plan step.
type autoApproveOnlySignal struct {
	autoApprove bool
}

func (*autoApproveOnlySignal) Type() signal.SignalType         { return "test-auto-approve-only" }
func (*autoApproveOnlySignal) Validate(workflow.Context) error { return nil }
func (*autoApproveOnlySignal) Execute(workflow.Context) error  { return nil }
func (s *autoApproveOnlySignal) AutoApproveOnPoliciesPassing(workflow.Context) bool {
	return s.autoApprove
}

type plainSignal struct{}

func (*plainSignal) Type() signal.SignalType         { return "test-plain" }
func (*plainSignal) Validate(workflow.Context) error { return nil }
func (*plainSignal) Execute(workflow.Context) error  { return nil }

func TestCheckShouldRun(t *testing.T) {
	checkCtx := &directive.CheckContext{}
	check := New(&policySignal{}, nil, checkCtx)
	step := &app.WorkflowStep{}

	require.False(t, check.ShouldRun(step, &app.Workflow{PlanOnly: true}))
	require.True(t, check.ShouldRun(step, &app.Workflow{}))

	checkCtx.NoopPlan = true
	require.False(t, check.ShouldRun(step, &app.Workflow{}))
}

func TestCheckShouldRunForAutoApproveWithoutPolicyEvaluation(t *testing.T) {
	step := &app.WorkflowStep{}

	checkCtx := &directive.CheckContext{}
	autoApproveCheck := New(&autoApproveOnlySignal{autoApprove: true}, nil, checkCtx)
	require.True(t, autoApproveCheck.ShouldRun(step, &app.Workflow{}))
	require.False(t, autoApproveCheck.ShouldRun(step, &app.Workflow{PlanOnly: true}))

	checkCtx.NoopPlan = true
	require.False(t, autoApproveCheck.ShouldRun(step, &app.Workflow{}))

	require.False(t, New(&plainSignal{}, nil, &directive.CheckContext{}).ShouldRun(step, &app.Workflow{}))
}

func runCheck(t *testing.T, sig signal.Signal) directive.CheckResult {
	t.Helper()

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))

	check := New(sig, nil, &directive.CheckContext{})
	env.ExecuteWorkflow(func(ctx workflow.Context) (directive.CheckResult, error) {
		return check.Run(ctx, &app.WorkflowStep{
			ID:             "step-1",
			StepTargetID:   "run-1",
			StepTargetType: "app_branch_runs",
		}, &app.Workflow{ID: "workflow-1", OrgID: "org-1"})
	})

	require.NoError(t, env.GetWorkflowError())
	var result directive.CheckResult
	require.NoError(t, env.GetWorkflowResult(&result))
	return result
}

// A signal with no policy target never evaluates policies, so its opt-in has to
// be honored on its own or the setting silently does nothing.
func TestCheckAutoApprovesSignalWithoutPolicyEvaluation(t *testing.T) {
	result := runCheck(t, &autoApproveOnlySignal{autoApprove: true})

	require.Equal(t, directive.StepContinue, result.Directive)
	require.Equal(t, app.WorkflowStepApprovalStatusApproved, result.Status)
	require.Equal(t, "policy-auto-approve", result.Reason.Check)
	require.Equal(t, "true", result.Reason.Labels["auto_approved"])
	require.Equal(t, "policies_passed", result.Reason.Labels["approval_reason"])
}

func TestCheckPassesWhenAutoApproveDisabled(t *testing.T) {
	result := runCheck(t, &autoApproveOnlySignal{autoApprove: false})

	require.Equal(t, directive.StepUnknown, result.Directive)
}

func TestCheckDoesNotAutoApproveWhenPolicyReportPersistenceFails(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))
	env.RegisterActivityWithOptions((&activities.Activities{}).RecordPolicyEvaluationCompositeError,
		activity.RegisterOptions{Name: "RecordPolicyEvaluationCompositeError"})

	env.OnActivity((*activities.Activities).PrepPolicyEvaluation, mock.Anything, mock.Anything, mock.Anything).
		Return(&activities.PrepPolicyEvaluationResult{
			HasPolicies: true,
			OrgID:       "org-1",
			AppID:       "app-1",
			PolicyIDs:   []string{"policy-1"},
			InputCount:  1,
			Policies: []activities.PolicyToEvaluate{{
				PolicyID:  "policy-1",
				Contents:  "package nuon\ndeny := []",
				InputJSON: []byte(`{}`),
			}},
		}, nil).Once()
	env.OnActivity((*activities.Activities).EvaluateSinglePolicy, mock.Anything, mock.Anything, mock.Anything).
		Return(&activities.EvaluateSinglePolicyResult{}, nil).Once()
	env.OnActivity((*activities.Activities).PersistPolicyReport, mock.Anything, mock.Anything, mock.Anything).
		Return((*activities.PersistPolicyReportResult)(nil), errors.New("write failed")).Times(3)
	env.OnActivity("RecordPolicyEvaluationCompositeError", mock.Anything, mock.MatchedBy(func(req activities.RecordPolicyEvaluationCompositeErrorRequest) bool {
		return req.Stage == policyerrors.EvaluationFailureStagePersistence
	})).Return(nil).Once()
	env.OnActivity((*statusactivities.Activities).PkgStatusUpdateFlowStepStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	check := New(&policySignal{}, nil, &directive.CheckContext{})
	env.ExecuteWorkflow(func(ctx workflow.Context) (directive.CheckResult, error) {
		return check.Run(ctx, &app.WorkflowStep{
			ID:             "step-1",
			StepTargetID:   "deploy-1",
			StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploy),
		}, &app.Workflow{ID: "workflow-1", OrgID: "org-1"})
	})

	require.NoError(t, env.GetWorkflowError())
	var result directive.CheckResult
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, directive.StepUnknown, result.Directive)
	env.AssertExpectations(t)
}
