package testworker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func executeComponentApply(ctx workflow.Context, sig componentdeployapplyplan.Signal) error {
	return sig.Execute(ctx)
}

func (e *FlowTestSuite) TestApplyRequiresAuthorizedPlan() {
	ns, err := e.service.TClient.GetNamespaceClient(defaultNamespace)
	require.NoError(e.T(), err)
	taskQueue := "apply-guard-" + fakeString()
	w := worker.New(ns, taskQueue, worker.Options{DisableRegistrationAliasing: true})
	acts := installactivities.New(installactivities.Params{DB: e.service.DB})
	w.RegisterWorkflow(executeComponentApply)
	w.RegisterActivity(acts.GetInstallForInstallComponent)
	w.RegisterActivity(acts.GetInstallDeployForApplyStep)
	w.RegisterActivityWithOptions(func(_ context.Context, req installactivities.GetLogStreamRequest) (*app.LogStream, error) {
		return nil, temporal.NewNonRetryableApplicationError("runner preparation reached: "+req.LogStreamID, "RunnerPreparationReached", nil)
	}, activity.RegisterOptions{Name: "GetLogStream"})
	require.NoError(e.T(), w.Start())
	defer w.Stop()

	for name, tc := range map[string]struct {
		status         app.Status
		response       app.WorkflowStepResponseType
		policyApproved bool
		directive      string
		planOnly       bool
		cancelled      bool
		wrongGroup     bool
		wrongAttempt   bool
		wrongComponent bool
		missingPlan    bool
		missingStep    bool
		retiredPlan    bool
		allowed        bool
	}{
		"policy error":                     {status: app.StatusError},
		"skipped with continue":            {status: app.StatusUserSkipped, directive: "continue"},
		"pending approval":                 {status: app.AwaitingApproval, directive: "continue"},
		"success without approval":         {status: app.StatusSuccess, directive: "continue"},
		"denied with continue":             {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeDeny},
		"skip response":                    {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeSkipCurrent},
		"approved but skipped":             {status: app.StatusUserSkipped, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove},
		"approved but stopped":             {status: app.StatusSuccess, directive: "stop", response: app.WorkflowStepApprovalResponseTypeApprove},
		"plan only":                        {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, planOnly: true},
		"cancelled":                        {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, cancelled: true},
		"wrong group":                      {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, wrongGroup: true},
		"wrong retry attempt":              {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, wrongAttempt: true},
		"wrong component":                  {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, wrongComponent: true},
		"missing plan":                     {missingPlan: true},
		"missing step context":             {missingStep: true},
		"superseded plan":                  {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, retiredPlan: true},
		"manual approval":                  {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, allowed: true},
		"approve all":                      {status: app.WorkflowStepApprovalStatusApproved, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, allowed: true},
		"policy auto approval":             {status: app.WorkflowStepApprovalStatusApproved, directive: "continue", policyApproved: true, allowed: true},
		"denial overrides policy metadata": {status: app.WorkflowStepApprovalStatusApproved, directive: "continue", policyApproved: true, response: app.WorkflowStepApprovalResponseTypeDeny},
	} {
		e.T().Run(name, func(t *testing.T) {
			e := &FlowTestSuite{t: t, service: e.service}
			ctx := e.service.Seed.EnsureAccount(t.Context(), t)
			ctx = e.service.Seed.EnsureOrg(ctx, t)
			deploy := e.seedDeniedPlan(ctx)
			var component app.InstallComponent
			require.NoError(t, e.service.DB.WithContext(ctx).First(&component, "id = ?", deploy.InstallComponentID).Error)
			flw := app.Workflow{OwnerID: component.InstallID, OwnerType: "installs", Type: "test_apply_guard", PlanOnly: tc.planOnly}
			if tc.cancelled {
				flw.Status.Status = app.StatusCancelled
			}
			require.NoError(t, e.service.DB.WithContext(ctx).Create(&flw).Error)
			require.NoError(t, e.service.DB.WithContext(ctx).Model(deploy).Update("install_workflow_id", flw.ID).Error)
			logStream := app.LogStream{OwnerID: deploy.ID, OwnerType: "install_deploys"}
			require.NoError(t, e.service.DB.WithContext(ctx).Create(&logStream).Error)
			steps := []app.WorkflowStep{
				{Name: "plan database", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeApproval,
					StepTargetID: deploy.ID, StepTargetType: "install_deploys", Status: app.CompositeStatus{Status: tc.status},
					ResultDirective: tc.directive, Retried: tc.retiredPlan},
				{Name: "apply database", Idx: 101, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem},
			}
			if tc.policyApproved {
				steps[0].Status.Metadata = map[string]any{"check": "policy-auto-approve", "check_label_approval_reason": "policies_passed"}
			}
			if tc.wrongGroup {
				steps[0].GroupIdx = 2
			}
			if tc.wrongAttempt {
				steps[0].GroupRetryIdx = 1
			}
			e.createTestSteps(ctx, &flw, steps)
			if tc.missingPlan {
				require.NoError(t, e.service.DB.WithContext(ctx).Delete(&steps[0]).Error)
			} else {
				approval := e.seedApproval(ctx, &steps[0])
				if tc.response != "" {
					require.NoError(t, e.service.DB.WithContext(ctx).Create(&app.WorkflowStepApprovalResponse{
						InstallWorkflowStepApprovalID: approval.ID, Type: tc.response,
					}).Error)
				}
			}
			if tc.allowed {
				decoy := app.InstallDeploy{
					InstallWorkflowID: &flw.ID, InstallComponentID: component.ID, ComponentBuildID: deploy.ComponentBuildID,
					Status: app.InstallDeployStatusPendingApproval, StatusDescription: "newer unapproved deploy",
				}
				require.NoError(t, e.service.DB.WithContext(ctx).Create(&decoy).Error)
			}
			sig := componentdeployapplyplan.Signal{InstallComponentID: component.ID, ComponentID: component.ComponentID, InstallID: component.InstallID}
			sig.SetStepContext(steps[1].ID, flw.ID)
			if tc.missingStep {
				sig.SetStepContext("", flw.ID)
			}
			if tc.wrongComponent {
				sig.ComponentID = fakeString()
			}
			run, err := ns.ExecuteWorkflow(ctx, tclient.StartWorkflowOptions{
				ID: "apply-guard-" + flw.ID, TaskQueue: taskQueue, WorkflowExecutionTimeout: 30 * time.Second,
			}, executeComponentApply, sig)
			require.NoError(t, err)
			err = run.Get(ctx, nil)
			require.Error(t, err)
			if tc.allowed {
				require.Contains(t, err.Error(), "runner preparation reached: "+logStream.ID)
			} else {
				require.NotContains(t, err.Error(), "RunnerPreparationReached")
				require.Contains(t, err.Error(), "unable to get install deploy")
			}
			var jobs int64
			require.NoError(t, e.service.DB.WithContext(ctx).Model(&app.RunnerJob{}).
				Where(app.RunnerJob{OwnerID: deploy.ID, Operation: app.RunnerJobOperationTypeApplyPlan}).Count(&jobs).Error)
			require.Zero(t, jobs)
		})
	}
}
