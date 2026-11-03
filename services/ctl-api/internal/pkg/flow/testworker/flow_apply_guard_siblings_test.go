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
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func executeTeardownApply(ctx workflow.Context, sig componentteardownapplyplan.Signal) error {
	return sig.Execute(ctx)
}

func executeProvisionSandboxApply(ctx workflow.Context, sig provisionsandboxapplyplan.Signal) error {
	return sig.Execute(ctx)
}

func executeReprovisionSandboxApply(ctx workflow.Context, sig reprovisionsandboxapplyplan.Signal) error {
	return sig.Execute(ctx)
}

func executeDeprovisionSandboxApply(ctx workflow.Context, sig deprovisionsandboxapplyplan.Signal) error {
	return sig.Execute(ctx)
}

func (e *FlowTestSuite) TestPlanSignalsSkipTheirGroup() {
	require.True(e.T(), (&componentteardownsyncandplan.Signal{}).SkipGroup())
	require.True(e.T(), (&provisionsandboxplan.Signal{}).SkipGroup())
	require.True(e.T(), (&reprovisionsandboxplan.Signal{}).SkipGroup())
	require.True(e.T(), (&deprovisionsandboxplan.Signal{}).SkipGroup())
}

func (e *FlowTestSuite) TestTeardownApplyRequiresAuthorizedPlan() {
	ns, err := e.service.TClient.GetNamespaceClient(defaultNamespace)
	require.NoError(e.T(), err)
	taskQueue := "teardown-apply-guard-" + fakeString()
	w := worker.New(ns, taskQueue, worker.Options{DisableRegistrationAliasing: true})
	acts := installactivities.New(installactivities.Params{DB: e.service.DB})
	w.RegisterWorkflow(executeTeardownApply)
	w.RegisterActivity(acts.GetInstallForInstallComponent)
	w.RegisterActivity(acts.GetInstallDeployForApplyStep)
	w.RegisterActivity(acts.GetApprovedInstallDeployForApplyStep)
	w.RegisterActivityWithOptions(func(context.Context, installactivities.UpdateInstallWorkflowStepTargetRequest) error {
		return nil
	}, activity.RegisterOptions{Name: "UpdateInstallWorkflowStepTarget"})
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
			deploy := e.seedApplyGuardDeploy(ctx)
			var component app.InstallComponent
			require.NoError(t, e.service.DB.WithContext(ctx).First(&component, "id = ?", deploy.InstallComponentID).Error)
			flw := app.Workflow{OwnerID: component.InstallID, OwnerType: "installs", Type: "test_teardown_apply_guard", PlanOnly: tc.planOnly}
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
			sig := componentteardownapplyplan.Signal{InstallComponentID: component.ID, ComponentID: component.ComponentID, InstallID: component.InstallID}
			sig.SetStepContext(steps[1].ID, flw.ID)
			if tc.missingStep {
				sig.SetStepContext("", flw.ID)
			}
			if tc.wrongComponent {
				sig.ComponentID = fakeString()
			}
			run, err := ns.ExecuteWorkflow(ctx, tclient.StartWorkflowOptions{
				ID: "teardown-apply-guard-" + flw.ID, TaskQueue: taskQueue, WorkflowExecutionTimeout: 30 * time.Second,
			}, executeTeardownApply, sig)
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

type sandboxApplyCase struct {
	status         app.Status
	response       app.WorkflowStepResponseType
	policyApproved bool
	directive      string
	planOnly       bool
	cancelled      bool
	wrongGroup     bool
	wrongAttempt   bool
	wrongInstall   bool
	missingPlan    bool
	missingStep    bool
	retiredPlan    bool
	allowed        bool
}

func sandboxApplyCases() map[string]sandboxApplyCase {
	return map[string]sandboxApplyCase{
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
		"wrong install":                    {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, wrongInstall: true},
		"missing plan":                     {missingPlan: true},
		"missing step context":             {missingStep: true},
		"superseded plan":                  {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, retiredPlan: true},
		"manual approval":                  {status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, allowed: true},
		"approve all":                      {status: app.WorkflowStepApprovalStatusApproved, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, allowed: true},
		"policy auto approval":             {status: app.WorkflowStepApprovalStatusApproved, directive: "continue", policyApproved: true, allowed: true},
		"denial overrides policy metadata": {status: app.WorkflowStepApprovalStatusApproved, directive: "continue", policyApproved: true, response: app.WorkflowStepApprovalResponseTypeDeny},
	}
}

func (e *FlowTestSuite) seedApplyGuardSandbox(ctx context.Context) (*app.Install, *app.InstallSandbox, *app.InstallSandboxRun) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	require.NoError(e.T(), err)
	tx := e.service.DB.WithContext(ctx).Begin()
	require.NoError(e.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(e.T(), tx.Exec("SET LOCAL session_replication_role = replica").Error)
	create := func(value any) {
		require.NoError(e.T(), tx.Omit(clause.Associations).Create(value).Error)
	}
	application := app.App{Name: "acme", OrgID: orgID}
	create(&application)
	appConfig := app.AppConfig{AppID: application.ID, StatusDescription: "test config"}
	create(&appConfig)
	install := app.Install{Name: "acme", AppID: application.ID, AppConfigID: appConfig.ID}
	create(&install)
	sandbox := app.InstallSandbox{InstallID: install.ID, Status: app.InstallSandboxStatusEmpty, StatusDescription: "empty"}
	create(&sandbox)
	run := app.InstallSandboxRun{
		InstallID: install.ID, RunType: app.SandboxRunTypeProvision,
		Status: app.SandboxRunPendingApproval, StatusDescription: "awaiting approval",
	}
	create(&run)
	require.NoError(e.T(), tx.Commit().Error)
	return &install, &sandbox, &run
}

func (e *FlowTestSuite) insertUnchecked(ctx context.Context, value any) {
	tx := e.service.DB.WithContext(ctx).Begin()
	require.NoError(e.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(e.T(), tx.Exec("SET LOCAL session_replication_role = replica").Error)
	require.NoError(e.T(), tx.Omit(clause.Associations).Create(value).Error)
	require.NoError(e.T(), tx.Commit().Error)
}

func (e *FlowTestSuite) TestSandboxApplyRequiresAuthorizedPlan() {
	ns, err := e.service.TClient.GetNamespaceClient(defaultNamespace)
	require.NoError(e.T(), err)
	taskQueue := "sandbox-apply-guard-" + fakeString()
	w := worker.New(ns, taskQueue, worker.Options{DisableRegistrationAliasing: true})
	acts := installactivities.New(installactivities.Params{DB: e.service.DB})
	w.RegisterWorkflow(executeProvisionSandboxApply)
	w.RegisterWorkflow(executeReprovisionSandboxApply)
	w.RegisterWorkflow(executeDeprovisionSandboxApply)
	w.RegisterActivity(acts.GetInstallForSandbox)
	w.RegisterActivity(acts.GetInstallSandboxRunForApplyStep)
	w.RegisterActivity(acts.GetApprovedInstallSandboxRunForApplyStep)
	w.RegisterActivityWithOptions(func(context.Context, installactivities.UpdateRunStatusRequest) error {
		return nil
	}, activity.RegisterOptions{Name: "UpdateRunStatus"})
	w.RegisterActivityWithOptions(func(context.Context, statusactivities.UpdateRunStatusV2Request) error {
		return nil
	}, activity.RegisterOptions{Name: "UpdateRunStatusV2"})
	w.RegisterActivityWithOptions(func(_ context.Context, req installactivities.GetLogStreamRequest) (*app.LogStream, error) {
		return nil, temporal.NewNonRetryableApplicationError("runner preparation reached: "+req.LogStreamID, "RunnerPreparationReached", nil)
	}, activity.RegisterOptions{Name: "GetLogStream"})
	require.NoError(e.T(), w.Start())
	defer w.Stop()

	start := func(ctx context.Context, id string, workflow any, arg any) error {
		run, err := ns.ExecuteWorkflow(ctx, tclient.StartWorkflowOptions{
			ID: id, TaskQueue: taskQueue, WorkflowExecutionTimeout: 30 * time.Second,
		}, workflow, arg)
		if err != nil {
			return err
		}
		return run.Get(ctx, nil)
	}

	for name, tc := range sandboxApplyCases() {
		e.T().Run("provision/"+name, func(t *testing.T) {
			e := &FlowTestSuite{t: t, service: e.service}
			ctx := e.service.Seed.EnsureAccount(t.Context(), t)
			ctx = e.service.Seed.EnsureOrg(ctx, t)
			install, sandbox, run := e.seedApplyGuardSandbox(ctx)
			flw := app.Workflow{OwnerID: install.ID, OwnerType: "installs", Type: "test_sandbox_apply_guard", PlanOnly: tc.planOnly}
			if tc.cancelled {
				flw.Status.Status = app.StatusCancelled
			}
			require.NoError(t, e.service.DB.WithContext(ctx).Create(&flw).Error)
			require.NoError(t, e.service.DB.WithContext(ctx).Model(run).Update("install_workflow_id", flw.ID).Error)
			logStream := app.LogStream{OwnerID: run.ID, OwnerType: "install_sandbox_runs"}
			require.NoError(t, e.service.DB.WithContext(ctx).Create(&logStream).Error)
			steps := []app.WorkflowStep{
				{Name: "plan sandbox", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeApproval,
					StepTargetID: run.ID, StepTargetType: string(app.WorkflowStepTargetTypeInstallSandboxRuns), Status: app.CompositeStatus{Status: tc.status},
					ResultDirective: tc.directive, Retried: tc.retiredPlan},
				{Name: "apply sandbox", Idx: 101, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem},
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
			if tc.wrongInstall {
				other := app.Install{Name: "other", AppID: install.AppID, AppConfigID: install.AppConfigID}
				e.insertUnchecked(ctx, &other)
				tx := e.service.DB.WithContext(ctx).Begin()
				require.NoError(t, tx.Error)
				defer tx.Rollback()
				require.NoError(t, tx.Exec("SET LOCAL session_replication_role = replica").Error)
				require.NoError(t, tx.Model(&app.InstallSandboxRun{}).Where("id = ?", run.ID).Update("install_id", other.ID).Error)
				require.NoError(t, tx.Commit().Error)
			}
			if tc.allowed {
				decoy := app.InstallSandboxRun{
					InstallID: install.ID, InstallWorkflowID: &flw.ID, RunType: app.SandboxRunTypeProvision,
					Status: app.SandboxRunPendingApproval, StatusDescription: "newer unapproved run",
					CreatedAt: time.Now().Add(time.Minute),
				}
				e.insertUnchecked(ctx, &decoy)
				require.NoError(t, e.service.DB.WithContext(ctx).Create(&app.LogStream{OwnerID: decoy.ID, OwnerType: "install_sandbox_runs"}).Error)
			}
			sig := provisionsandboxapplyplan.Signal{InstallSandboxID: sandbox.ID, InstallID: install.ID}
			sig.SetStepContext(steps[1].ID, flw.ID)
			if tc.missingStep {
				sig.SetStepContext("", flw.ID)
			}
			err := start(ctx, "sandbox-apply-guard-"+flw.ID, executeProvisionSandboxApply, sig)
			require.Error(t, err)
			if tc.allowed {
				require.Contains(t, err.Error(), "runner preparation reached: "+logStream.ID)
			} else {
				require.NotContains(t, err.Error(), "RunnerPreparationReached")
				require.Contains(t, err.Error(), "unable to get install deploy")
			}
		})
	}

	for _, sibling := range []struct {
		name     string
		workflow any
		arg      func(sandboxID, installID, stepID, flowID string) any
	}{
		{
			name:     "reprovision",
			workflow: executeReprovisionSandboxApply,
			arg: func(sandboxID, installID, stepID, flowID string) any {
				sig := reprovisionsandboxapplyplan.Signal{InstallSandboxID: sandboxID, InstallID: installID}
				sig.SetStepContext(stepID, flowID)
				return sig
			},
		},
		{
			name:     "deprovision",
			workflow: executeDeprovisionSandboxApply,
			arg: func(sandboxID, installID, stepID, flowID string) any {
				sig := deprovisionsandboxapplyplan.Signal{InstallSandboxID: sandboxID, InstallID: installID}
				sig.SetStepContext(stepID, flowID)
				return sig
			},
		},
	} {
		for _, tc := range []sandboxApplyCase{
			{status: app.StatusUserSkipped, directive: "continue"},
			{status: app.StatusSuccess, directive: "continue", response: app.WorkflowStepApprovalResponseTypeApprove, allowed: true},
		} {
			name := sibling.name + "/skipped"
			if tc.allowed {
				name = sibling.name + "/manual approval"
			}
			e.T().Run(name, func(t *testing.T) {
				e := &FlowTestSuite{t: t, service: e.service}
				ctx := e.service.Seed.EnsureAccount(t.Context(), t)
				ctx = e.service.Seed.EnsureOrg(ctx, t)
				install, sandbox, run := e.seedApplyGuardSandbox(ctx)
				flw := app.Workflow{OwnerID: install.ID, OwnerType: "installs", Type: app.WorkflowType("test_sandbox_apply_guard_" + sibling.name)}
				require.NoError(t, e.service.DB.WithContext(ctx).Create(&flw).Error)
				require.NoError(t, e.service.DB.WithContext(ctx).Model(run).Update("install_workflow_id", flw.ID).Error)
				logStream := app.LogStream{OwnerID: run.ID, OwnerType: "install_sandbox_runs"}
				require.NoError(t, e.service.DB.WithContext(ctx).Create(&logStream).Error)
				steps := []app.WorkflowStep{
					{Name: "plan sandbox", Idx: 100, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeApproval,
						StepTargetID: run.ID, StepTargetType: string(app.WorkflowStepTargetTypeInstallSandboxRuns), Status: app.CompositeStatus{Status: tc.status},
						ResultDirective: tc.directive},
					{Name: "apply sandbox", Idx: 101, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem},
				}
				e.createTestSteps(ctx, &flw, steps)
				approval := e.seedApproval(ctx, &steps[0])
				if tc.response != "" {
					require.NoError(t, e.service.DB.WithContext(ctx).Create(&app.WorkflowStepApprovalResponse{
						InstallWorkflowStepApprovalID: approval.ID, Type: tc.response,
					}).Error)
				}
				if tc.allowed {
					decoy := app.InstallSandboxRun{
						InstallID: install.ID, InstallWorkflowID: &flw.ID, RunType: app.SandboxRunTypeProvision,
						Status: app.SandboxRunPendingApproval, StatusDescription: "newer unapproved run",
						CreatedAt: time.Now().Add(time.Minute),
					}
					e.insertUnchecked(ctx, &decoy)
					require.NoError(t, e.service.DB.WithContext(ctx).Create(&app.LogStream{OwnerID: decoy.ID, OwnerType: "install_sandbox_runs"}).Error)
				}
				err := start(ctx, sibling.name+"-apply-guard-"+flw.ID, sibling.workflow, sibling.arg(sandbox.ID, install.ID, steps[1].ID, flw.ID))
				require.Error(t, err)
				if tc.allowed {
					require.Contains(t, err.Error(), "runner preparation reached: "+logStream.ID)
				} else {
					require.NotContains(t, err.Error(), "RunnerPreparationReached")
					require.Contains(t, err.Error(), "unable to get install deploy")
				}
			})
		}
	}
}
