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

func TestCheckShouldRun(t *testing.T) {
	checkCtx := &directive.CheckContext{}
	check := New(&policySignal{}, nil, checkCtx)
	step := &app.WorkflowStep{}

	require.False(t, check.ShouldRun(step, &app.Workflow{PlanOnly: true}))
	require.True(t, check.ShouldRun(step, &app.Workflow{}))

	checkCtx.NoopPlan = true
	require.False(t, check.ShouldRun(step, &app.Workflow{}))
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
