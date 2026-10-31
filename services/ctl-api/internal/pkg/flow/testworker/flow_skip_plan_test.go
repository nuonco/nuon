package testworker

import (
	"bytes"
	"compress/gzip"
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (e *FlowTestSuite) seedDeniedPlan(ctx context.Context) *app.InstallDeploy {
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
	component := app.Component{Name: "database", AppID: application.ID, Type: app.ComponentTypeTerraformModule}
	create(&component)
	installedComponent := app.InstallComponent{InstallID: install.ID, ComponentID: component.ID}
	create(&installedComponent)
	build := app.ComponentBuild{ComponentID: component.ID}
	create(&build)
	deploy := app.InstallDeploy{
		InstallComponentID: installedComponent.ID, ComponentBuildID: build.ID,
		Status: app.InstallDeployStatusPending, StatusDescription: "awaiting plan",
	}
	create(&deploy)
	policies := app.AppPoliciesConfig{AppID: application.ID, AppConfigID: appConfig.ID}
	create(&policies)
	create(&app.AppPolicyConfig{
		AppID: application.ID, AppConfigID: appConfig.ID, AppPoliciesConfigID: policies.ID,
		Name: "block-destructive-changes", Type: config.AppPolicyTypeTerraformModule, Engine: config.AppPolicyEngineOPA,
		Components: []string{component.Name},
		Contents: `package nuon
deny contains "RDS replacement is forbidden" if {
  some change in input.plan.resource_changes
  change.type == "aws_db_instance"
  "delete" in change.change.actions
}`,
	})
	job := app.RunnerJob{
		OwnerID: deploy.ID, OwnerType: "install_deploys",
		Type: app.RunnerJobTypeTerraformDeploy, Operation: app.RunnerJobOperationTypeCreateApplyPlan,
		Status: app.RunnerJobStatusFinished, StatusDescription: "plan generated", MaxExecutions: 1,
		QueueTimeout: time.Minute, AvailableTimeout: time.Minute, ExecutionTimeout: time.Minute,
	}
	create(&job)
	execution := app.RunnerJobExecution{RunnerJobID: job.ID, Status: app.RunnerJobExecutionStatusFinished}
	create(&execution)
	var plan bytes.Buffer
	zw := gzip.NewWriter(&plan)
	_, err = zw.Write([]byte(`{"format_version":"1.2","terraform_version":"1.9.8","resource_changes":[{"address":"aws_db_instance.database","mode":"managed","type":"aws_db_instance","name":"database","provider_name":"registry.terraform.io/hashicorp/aws","change":{"actions":["delete","create"],"before":{"storage_encrypted":false},"after":{"storage_encrypted":true},"replace_paths":[["storage_encrypted"]]}}]}`))
	require.NoError(e.T(), err)
	require.NoError(e.T(), zw.Close())
	create(&app.RunnerJobExecutionResult{RunnerJobExecutionID: execution.ID, ContentsDisplayGzip: plan.Bytes()})
	require.NoError(e.T(), tx.Commit().Error)
	return &deploy
}

func (e *FlowTestSuite) TestSkipPolicyDeniedPlanDiscardsApply() {
	e.testSkipPlanDiscardsApply(true, true)
}

func (e *FlowTestSuite) TestSkipAwaitingApprovalPlanDiscardsApply() {
	e.testSkipPlanDiscardsApply(false, true)
}

func (e *FlowTestSuite) TestSkipLegacyAwaitingApprovalPlanDiscardsApply() {
	e.testSkipPlanDiscardsApply(false, false)
}

func (e *FlowTestSuite) TestSkipSupersededPlanPreservesReplacement() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("failed plan", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		approvalStep("replacement plan", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{}}),
		{Name: "apply database", Idx: 102, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	}
	steps[0].Retried = true
	steps[0].Status.Status = app.StatusError
	steps[1].Idx = 101
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	approval := e.seedApproval(ctx, &steps[1])
	e.enqueueLifecycleFlow(ctx, queueID, flw, ownerID, ownerType)
	e.awaitApprovalParked(ctx, flw, steps[1].ID)
	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID, StepID: steps[0].ID,
	})
	require.NoError(e.T(), err)
	require.False(e.T(), resp.Skippable)
	require.Equal(e.T(), app.StatusError, e.getStep(ctx, steps[0].ID).Status.Status)
	require.Equal(e.T(), app.AwaitingApproval, e.getStep(ctx, steps[1].ID).Status.Status)
	require.Equal(e.T(), app.StatusPending, e.getStep(ctx, steps[2].ID).Status.Status)
	e.respondApproval(ctx, flw, &steps[1], approval.ID, app.WorkflowStepApprovalResponseTypeApprove)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusSuccess, e.getStep(ctx, steps[2].ID).Status.Status)
}

func (e *FlowTestSuite) testSkipPlanDiscardsApply(policyDenied, resident bool) {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := newTestOwner()
	steps := []app.WorkflowStep{
		approvalStep("plan database", 1, signaldb.SignalData{Signal: &SkipGroupApprovalSignal{EvaluatePolicies: policyDenied}}),
		{Name: "apply database", Idx: 101, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "post-deploy action", Idx: 102, GroupIdx: 1, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
		{Name: "independent component", Idx: 200, GroupIdx: 2, ExecutionType: app.WorkflowStepExecutionTypeSystem,
			QueueSignal: &signaldb.SignalData{Signal: &SuccessSignal{}}},
	}
	if policyDenied {
		deploy := e.seedDeniedPlan(ctx)
		steps[0].StepTargetID = deploy.ID
		steps[0].StepTargetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	}
	flw, queueID := e.setupLifecycleTest(ctx, ownerID, ownerType, steps)
	e.seedApproval(ctx, &steps[0])
	_, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     queuenames.InstallWorkflowStepGroupsQueueName,
			StepQueueName:          queuenames.InstallWorkflowStepsQueueName,
			StepTargetQueueName:    queuenames.InstallSignalsQueueName,
			GenerateStepsQueueName: queuenames.InstallGenerateStepsQueueName,
			OwnerID:                ownerID,
			OwnerType:              ownerType,
			Resident:               resident,
			ResidentIdleTimeout:    testResidentIdleTimeout,
		},
		OwnerID: flw.ID, OwnerType: "install_workflows",
	})
	require.NoError(e.T(), err)
	if policyDenied {
		e.waitForWorkflowStatus(ctx, flw.ID, app.StatusError)
		plan := e.getStep(ctx, steps[0].ID)
		require.Equal(e.T(), app.StatusError, plan.Status.Status)
		require.Equal(e.T(), "policy", plan.Status.Metadata["check"])
		require.NotEmpty(e.T(), plan.Status.Metadata["deny_violations"])
		require.Empty(e.T(), plan.ResultDirective)
		e.assertTemporalDrained(ctx, flw.ID)
	} else {
		e.awaitApprovalParked(ctx, flw, steps[0].ID)
	}
	for _, step := range steps[1:] {
		require.Equal(e.T(), app.StatusPending, e.getStep(ctx, step.ID).Status.Status)
	}
	resp, err := e.service.FlowClient.SkipStep(ctx, &flowclient.SkipStepRequest{
		InstallWorkflowID: flw.ID, StepID: steps[0].ID,
	})
	require.NoError(e.T(), err)
	require.True(e.T(), resp.Skippable)
	e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
	require.Equal(e.T(), app.StatusUserSkipped, e.getStep(ctx, steps[0].ID).Status.Status)
	for _, step := range steps[1:3] {
		actual := e.getStep(ctx, step.ID)
		require.Equal(e.T(), app.StatusDiscarded, actual.Status.Status, step.Name)
		require.True(e.T(), actual.StartedAt.IsZero(), "%s must never start", step.Name)
		var dispatched int64
		require.NoError(e.T(), e.service.DB.WithContext(ctx).Model(&app.QueueSignal{}).
			Where(app.QueueSignal{OwnerID: step.ID}).Count(&dispatched).Error)
		require.Zero(e.T(), dispatched, "%s must never be dispatched", step.Name)
	}
	require.Equal(e.T(), app.StatusSuccess, e.getStep(ctx, steps[3].ID).Status.Status)
}
