package policy

import (
	"context"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	env.OnActivity((&activities.Activities{}).PrepPolicyEvaluation, mock.Anything, mock.Anything).
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
	env.OnActivity((&activities.Activities{}).EvaluateSinglePolicy, mock.Anything, mock.Anything).
		Return(&activities.EvaluateSinglePolicyResult{}, nil).Once()
	env.OnActivity((&activities.Activities{}).PersistPolicyReport, mock.Anything, mock.Anything).
		Return((*activities.PersistPolicyReportResult)(nil), errors.New("write failed")).Times(3)

	check := New(&policySignal{}, nil, &directive.CheckContext{})
	env.ExecuteWorkflow(func(ctx workflow.Context) (directive.CheckResult, error) {
		return check.Run(ctx, &app.WorkflowStep{
			ID:             "step-1",
			StepTargetID:   "deploy-1",
			StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploy),
		}, &app.Workflow{ID: "workflow-1", OrgID: "org-1"})
	})

	require.ErrorContains(t, env.GetWorkflowError(), "unable to persist policy report")
	env.AssertExpectations(t)
}

func TestCheckNamedPolicyReportingAndEnforcement(t *testing.T) {
	const warningPolicy = `# METADATA
# custom:
#   action: warn
package nuon.checks
import data.nuon.lib.check
# METADATA
# title: Review changes
check_shared contains check.finding("review changes") if { json.unmarshal(input.document) }`
	const passingPolicy = `package nuon.checks
import data.nuon.lib.check
# METADATA
# title: Restricted changes
check_shared contains check.finding("blocked") if { input.blocked }`
	for name, tt := range map[string]struct {
		secondInput    string
		secondDecision string
		blocked        bool
	}{
		"warnings allow approval": {`{"document":"false"}`, "pass", false},
		"errors block approval":   {`{"document":"invalid"}`, "error", true},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
			env.SetDataConverter(converter.NewCompositeDataConverter(
				converter.NewNilPayloadConverter(),
				converter.NewByteSlicePayloadConverter(),
				pkgdataconverter.NewJSONConverter(),
			))
			env.RegisterActivity((&activities.Activities{}).EvaluateSinglePolicy)
			env.RegisterActivityWithOptions(func(context.Context, any) error { return nil },
				activity.RegisterOptions{Name: "PkgStatusUpdateFlowStepStatus"})
			prep := &activities.PrepPolicyEvaluationResult{
				HasPolicies: true, OrgID: "org-1", AppID: "app-1",
				PolicyIDs: []string{"policy-a", "policy-b"}, InputCount: 2,
			}
			for i, input := range []string{`{"document":"true"}`, tt.secondInput} {
				for _, policy := range []struct{ id, contents string }{
					{"policy-a", warningPolicy}, {"policy-b", passingPolicy},
				} {
					prep.Policies = append(prep.Policies, activities.PolicyToEvaluate{
						PolicyID: policy.id, Contents: policy.contents,
						InputJSON: []byte(input), InputIndex: i,
						InputIdentity: []string{"Deployment/default/first", "Deployment/default/second"}[i],
					})
				}
			}
			env.OnActivity((&activities.Activities{}).PrepPolicyEvaluation, mock.Anything, mock.Anything).
				Return(prep, nil)
			reports := make(chan *activities.PersistPolicyReportRequest, 1)
			env.OnActivity((&activities.Activities{}).PersistPolicyReport, mock.Anything, mock.Anything).
				Return(func(_ context.Context, req *activities.PersistPolicyReportRequest) (*activities.PersistPolicyReportResult, error) {
					reports <- req
					return &activities.PersistPolicyReportResult{ReportID: "report-1", PassedPolicyIDs: []string{"policy-b"}}, nil
				})

			check := New(&policySignal{}, nil, &directive.CheckContext{})
			env.ExecuteWorkflow(func(ctx workflow.Context) (directive.CheckResult, error) {
				return check.Run(ctx, &app.WorkflowStep{
					ID: "step-1", StepTargetID: "deploy-1",
					StepTargetType: string(app.WorkflowStepTargetTypeInstallDeploy),
				}, &app.Workflow{ID: "workflow-1", OrgID: "org-1"})
			})
			if tt.blocked {
				require.ErrorContains(t, env.GetWorkflowError(), "policy violations found")
			} else {
				require.NoError(t, env.GetWorkflowError())
				var result directive.CheckResult
				require.NoError(t, env.GetWorkflowResult(&result))
				require.Equal(t, directive.StepContinue, result.Directive)
				require.Equal(t, app.WorkflowStepApprovalStatusApproved, result.Status)
			}
			var report *activities.PersistPolicyReportRequest
			select {
			case report = <-reports:
			default:
				t.Fatal("workflow did not submit a policy report")
			}
			require.Equal(t, map[string]int{"policy-a": 2, "policy-b": 2}, report.PolicyInputCounts)
			require.Len(t, report.CheckResults, 2)
			for policyID, title := range map[string]string{"policy-a": "Review changes", "policy-b": "Restricted changes"} {
				checks := report.CheckResults[policyID]
				require.Len(t, checks, 2)
				decisions := make([]string, 0, 2)
				findings := make(map[int]string)
				evaluationErrors := make(map[int]string)
				for _, result := range checks {
					require.Equal(t, "check_shared", result.CheckID)
					require.Equal(t, title, result.Title)
					require.Equal(t, 1, result.InputCount)
					decisions = append(decisions, result.Decision)
					for _, finding := range result.Findings {
						findings[finding.InputIndex] = finding.Message
						require.Equal(t, "Deployment/default/first", finding.InputIdentity)
					}
					for _, evaluationErr := range result.Errors {
						evaluationErrors[evaluationErr.InputIndex] = evaluationErr.Code
						require.Equal(t, "Deployment/default/second", evaluationErr.InputIdentity)
					}
				}
				if policyID == "policy-a" {
					require.ElementsMatch(t, []string{"warn", tt.secondDecision}, decisions)
					require.Equal(t, map[int]string{0: "review changes"}, findings)
				} else {
					require.Equal(t, []string{"pass", "pass"}, decisions)
					require.Empty(t, findings)
				}
				if policyID == "policy-a" && tt.blocked {
					require.Equal(t, map[int]string{1: "evaluation_failed"}, evaluationErrors)
				} else {
					require.Empty(t, evaluationErrors)
				}
			}
			severities := make(map[int]string)
			for _, violation := range report.Violations {
				require.Equal(t, "policy-a", violation.PolicyID)
				require.Equal(t, "check_shared", violation.CheckID)
				require.Equal(t, []string{"Deployment/default/first", "Deployment/default/second"}[violation.InputIndex], violation.InputIdentity)
				severities[violation.InputIndex] = violation.Severity
			}
			expected := map[int]string{0: "warn"}
			if tt.blocked {
				expected[1] = "deny"
			}
			require.Len(t, report.Violations, len(expected))
			require.Equal(t, expected, severities)
		})
	}
}
