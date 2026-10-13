package activities_test

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func (s *InstallStateIntegrationSuite) componentConnection() app.ComponentConfigConnection {
	s.T().Helper()
	var connection app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Preload("Component").
		Where(app.ComponentConfigConnection{AppConfigID: s.install.AppConfigID, Type: app.ComponentTypeTerraformModule}).
		First(&connection).Error)
	return connection
}

func (s *InstallStateIntegrationSuite) create(objs ...any) {
	s.T().Helper()
	for _, obj := range objs {
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(obj).Error)
	}
}

func (s *InstallStateIntegrationSuite) createRunner(status app.RunnerStatus) *app.Runner {
	group := &app.RunnerGroup{
		ID: domains.NewRunnerGroupID(), OrgID: s.install.OrgID, OwnerID: s.install.ID, OwnerType: "installs",
		Type: app.RunnerGroupTypeInstall, Platform: app.AppRunnerTypeAWS,
	}
	runner := &app.Runner{
		ID: domains.NewRunnerID(), OrgID: s.install.OrgID, Name: "example-runner", DisplayName: "Example Runner",
		Status: status, RunnerGroupID: group.ID,
	}
	s.create(group, runner)
	return runner
}

func (s *InstallStateIntegrationSuite) createExecution(runnerID, ownerID, ownerType string, jobType app.RunnerJobType, status app.RunnerJobExecutionStatus) *app.RunnerJobExecution {
	logStream := &app.LogStream{ID: domains.NewLogStreamID(), OrgID: s.install.OrgID, OwnerID: ownerID, OwnerType: ownerType}
	s.create(logStream)
	job := &app.RunnerJob{
		LogStreamID: &logStream.ID, RunnerID: runnerID, OwnerID: ownerID, OwnerType: ownerType, Type: jobType,
		Status: app.RunnerJobStatusFinished, Group: app.RunnerJobGroupDeploy,
		Operation: app.RunnerJobOperationTypeApplyPlan, MaxExecutions: 1,
	}
	s.create(job)
	execution := &app.RunnerJobExecution{RunnerJobID: job.ID, Status: status}
	s.create(execution)
	return execution
}

func executionOutputs(executionID string, outputs map[string]any) *app.RunnerJobExecutionOutputs {
	encoded, err := json.Marshal(outputs)
	if err != nil {
		panic(err)
	}
	return &app.RunnerJobExecutionOutputs{RunnerJobExecutionID: executionID, Outputs: encoded}
}

func (s *InstallStateIntegrationSuite) createDeploy(installComponentID, connectionID string, status app.InstallDeployStatus) *app.InstallDeploy {
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), connectionID)
	deploy := &app.InstallDeploy{
		InstallComponentID: installComponentID, ComponentBuildID: build.ID,
		Status: status, Type: app.InstallDeployTypeApply,
	}
	s.create(deploy)
	return deploy
}

func componentState(t require.TestingT, snapshot *state.State, name string) map[string]any {
	component, ok := snapshot.Components[name].(map[string]any)
	require.True(t, ok, "component %s missing from %v", name, snapshot.Components)
	return component
}

func (s *InstallStateIntegrationSuite) TestSourceCommitsReachStrictFreshReads() {
	connection := s.componentConnection()
	definition := s.deps.Seeder.CreateActionWorkflow(s.ctx, s.T(), s.install.AppID)
	s.deps.Seeder.CreateActionWorkflowConfig(s.ctx, s.T(), s.install.AppID, s.install.AppConfigID, definition.ID)
	var (
		runner          *app.Runner
		deploy          *app.InstallDeploy
		deployExec      *app.RunnerJobExecution
		sandboxExec     *app.RunnerJobExecution
		installedAction *app.InstallActionWorkflow
		actionExec      *app.RunnerJobExecution
		secretsExec     *app.RunnerJobExecution
	)
	for _, step := range []struct {
		name   string
		commit func()
		check  func(*state.State)
	}{
		{
			name:   "runner_created",
			commit: func() { runner = s.createRunner(app.RunnerStatusProvisioning) },
			check: func(snapshot *state.State) {
				require.Equal(s.T(), runner.ID, snapshot.Runner.ID)
				require.Equal(s.T(), string(app.RunnerStatusProvisioning), snapshot.Runner.Status)
			},
		},
		{
			name: "runner_status",
			commit: func() {
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(runner).Update("status", app.RunnerStatusActive).Error)
			},
			check: func(snapshot *state.State) {
				require.Equal(s.T(), string(app.RunnerStatusActive), snapshot.Runner.Status)
			},
		},
		{
			name: "component_deploy",
			commit: func() {
				installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), s.install.ID, connection.ComponentID)
				deploy = s.createDeploy(installed.ID, connection.ID, app.InstallDeployStatusActive)
				deployExec = s.createExecution(runner.ID, deploy.ID, "install_deploys", app.RunnerJobTypeTerraformDeploy, app.RunnerJobExecutionStatusFinished)
			},
			check: func(snapshot *state.State) {
				component := componentState(s.T(), snapshot, connection.Component.Name)
				require.Equal(s.T(), string(app.InstallDeployStatusActive), component["status"])
				require.Equal(s.T(), deploy.ComponentBuildID, component["build_id"])
				require.Empty(s.T(), component["outputs"])
			},
		},
		{
			name:   "component_outputs",
			commit: func() { s.create(executionOutputs(deployExec.ID, map[string]any{"endpoint": "api.example.com"})) },
			check: func(snapshot *state.State) {
				component := componentState(s.T(), snapshot, connection.Component.Name)
				require.Equal(s.T(), map[string]any{"endpoint": "api.example.com"}, component["outputs"])
			},
		},
		{
			name: "sandbox_run",
			commit: func() {
				run := &app.InstallSandboxRun{
					InstallID: s.install.ID, AppSandboxConfigID: s.install.AppSandboxConfigID,
					RunType: app.SandboxRunTypeProvision, Status: app.SandboxRunStatusActive, StatusDescription: "active",
				}
				s.create(run)
				sandboxExec = s.createExecution(runner.ID, run.ID, "install_sandbox_runs", app.RunnerJobTypeSandboxTerraform, app.RunnerJobExecutionStatusFinished)
			},
			check: func(snapshot *state.State) {
				require.Equal(s.T(), string(app.SandboxRunStatusActive), snapshot.Sandbox.Status)
				require.Empty(s.T(), snapshot.Domain.PublicDomain)
			},
		},
		{
			name: "sandbox_outputs",
			commit: func() {
				s.create(executionOutputs(sandboxExec.ID, map[string]any{"public_domain": "acme.example.com", "internal_domain": "acme.internal"}))
			},
			check: func(snapshot *state.State) {
				require.Equal(s.T(), "acme.example.com", snapshot.Sandbox.Outputs["public_domain"])
				require.Equal(s.T(), "acme.example.com", snapshot.Domain.PublicDomain)
				require.Equal(s.T(), "acme.internal", snapshot.Install.InternalDomain)
			},
		},
		{
			name: "action_installed",
			commit: func() {
				installedAction = &app.InstallActionWorkflow{InstallID: s.install.ID, ActionWorkflowID: definition.ID}
				s.create(installedAction)
			},
			check: func(snapshot *state.State) {
				require.Equal(s.T(), string(app.InstallActionRunStatusUnknown), snapshot.Actions.Workflows[definition.Name].Status)
			},
		},
		{
			name: "action_run_finished",
			commit: func() {
				run := &app.InstallActionWorkflowRun{
					InstallID: s.install.ID, InstallActionWorkflowID: generics.NewNullString(installedAction.ID),
					Status: app.InstallActionRunStatusFinished, StatusDescription: "finished",
					TriggerType: app.ActionWorkflowTriggerTypeManual,
				}
				s.create(run)
				actionExec = s.createExecution(runner.ID, run.ID, "install_action_workflow_runs", app.RunnerJobTypeTerraformDeploy, app.RunnerJobExecutionStatusFinished)
			},
			check: func(snapshot *state.State) {
				require.Equal(s.T(), string(app.InstallActionRunStatusFinished), snapshot.Actions.Workflows[definition.Name].Status)
				require.Empty(s.T(), snapshot.Actions.Workflows[definition.Name].Outputs)
			},
		},
		{
			name:   "action_outputs",
			commit: func() { s.create(executionOutputs(actionExec.ID, map[string]any{"token_id": "example"})) },
			check: func(snapshot *state.State) {
				require.Equal(s.T(), "example", snapshot.Actions.Workflows[definition.Name].Outputs["token_id"])
			},
		},
		{
			name: "stack_version",
			commit: func() {
				s.deps.Seeder.CreateInstallStackVersion(s.ctx, s.T(), s.install.ID, s.install.InstallStack.ID, s.install.AppConfigID)
			},
			check: func(snapshot *state.State) {
				require.NotNil(s.T(), snapshot.InstallStack)
				require.Equal(s.T(), "test-checksum", snapshot.InstallStack.Checksum)
				require.Empty(s.T(), snapshot.InstallStack.Outputs)
			},
		},
		{
			name: "stack_outputs",
			commit: func() {
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallStackOutputs{}).
					Where(app.InstallStackOutputs{InstallStackID: s.install.InstallStack.ID}).
					Update("data", pgtype.Hstore{"vpc_id": generics.ToPtr("vpc-0example")}).Error)
			},
			check: func(snapshot *state.State) {
				encoded, err := json.Marshal(snapshot.InstallStack.Outputs)
				require.NoError(s.T(), err)
				require.JSONEq(s.T(), `{"vpc_id":"vpc-0example"}`, string(encoded))
			},
		},
		{
			name: "secrets_sync_job",
			commit: func() {
				secretsExec = s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeSandboxSyncSecrets, app.RunnerJobExecutionStatusFinished)
			},
			check: func(snapshot *state.State) {
				require.NotNil(s.T(), snapshot.Secrets)
				require.Empty(s.T(), *snapshot.Secrets)
			},
		},
		{
			name: "secrets_sync_outputs",
			commit: func() {
				s.create(executionOutputs(secretsExec.ID, map[string]any{"example": map[string]any{"name": "example", "exists": true}}))
			},
			check: func(snapshot *state.State) {
				require.True(s.T(), (*snapshot.Secrets)["example"].Exists)
			},
		},
	} {
		s.Run(step.name, func() {
			step.commit()
			for range 2 {
				step.check(s.fresh())
			}
		})
	}
}

func (s *InstallStateIntegrationSuite) raceSourceCommits(first, second func(*gorm.DB) error) {
	tx := s.deps.DB.WithContext(s.ctx).Begin()
	require.NoError(s.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(s.T(), first(tx))
	pid := s.backendPID(tx)

	done := make(chan struct{})
	var secondErr error
	go func() {
		defer close(done)
		secondErr = s.deps.DB.WithContext(s.ctx).Transaction(second)
	}()
	if s.blockedBy(pid, done) {
		s.T().Log("second source transaction waited for the first transaction's head invalidation")
	} else {
		require.NoError(s.T(), secondErr)
		s.fresh()
	}
	require.NoError(s.T(), tx.Commit().Error)
	s.await(done)
	require.NoError(s.T(), secondErr)
}

func (s *InstallStateIntegrationSuite) TestOutputRacesInvalidateAfterBothCommits() {
	connection := s.componentConnection()
	runner := s.createRunner(app.RunnerStatusActive)
	installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), s.install.ID, connection.ComponentID)
	for _, tc := range []struct {
		name      string
		deploy    app.InstallDeployStatus
		execution app.RunnerJobExecutionStatus
		endpoint  string
		first     func(tx *gorm.DB, execution *app.RunnerJobExecution, outputs *app.RunnerJobExecutionOutputs) error
		second    func(tx *gorm.DB, deploy *app.InstallDeploy, execution *app.RunnerJobExecution) error
	}{
		{
			name: "outputs_begun_while_deploy_not_active", deploy: app.InstallDeployStatusPlanning,
			execution: app.RunnerJobExecutionStatusFinished, endpoint: "activated.example.com",
			first: func(tx *gorm.DB, _ *app.RunnerJobExecution, outputs *app.RunnerJobExecutionOutputs) error {
				return tx.Create(outputs).Error
			},
			second: func(tx *gorm.DB, deploy *app.InstallDeploy, _ *app.RunnerJobExecution) error {
				return tx.Model(deploy).Update("status", app.InstallDeployStatusActive).Error
			},
		},
		{
			name: "execution_completes_before_outputs_visible", deploy: app.InstallDeployStatusActive,
			execution: app.RunnerJobExecutionStatusInProgress, endpoint: "completed.example.com",
			first: func(tx *gorm.DB, execution *app.RunnerJobExecution, _ *app.RunnerJobExecutionOutputs) error {
				return tx.Model(execution).Update("status", app.RunnerJobExecutionStatusFinished).Error
			},
			second: func(tx *gorm.DB, _ *app.InstallDeploy, execution *app.RunnerJobExecution) error {
				return tx.Create(executionOutputs(execution.ID, map[string]any{"endpoint": "completed.example.com"})).Error
			},
		},
	} {
		s.Run(tc.name, func() {
			deploy := s.createDeploy(installed.ID, connection.ID, tc.deploy)
			execution := s.createExecution(runner.ID, deploy.ID, "install_deploys", app.RunnerJobTypeTerraformDeploy, app.RunnerJobExecutionStatusFinished)
			if tc.execution != app.RunnerJobExecutionStatusFinished {
				s.create(executionOutputs(execution.ID, map[string]any{"endpoint": "previous.example.com"}))
				execution = &app.RunnerJobExecution{RunnerJobID: execution.RunnerJobID, Status: tc.execution}
				s.create(execution)
			}
			s.fresh()
			s.raceSourceCommits(
				func(tx *gorm.DB) error {
					return tc.first(tx, execution, executionOutputs(execution.ID, map[string]any{"endpoint": tc.endpoint}))
				},
				func(tx *gorm.DB) error { return tc.second(tx, deploy, execution) },
			)
			for range 2 {
				component := componentState(s.T(), s.fresh(), connection.Component.Name)
				require.Equal(s.T(), deploy.ComponentBuildID, component["build_id"])
				require.Equal(s.T(), map[string]any{"endpoint": tc.endpoint}, component["outputs"])
			}
		})
	}
}

func (s *InstallStateIntegrationSuite) TestRegenerateWorkflowRepublishesFromCurrentHead() {
	require.NoError(s.T(), s.renameApp("before-upload"))
	var once sync.Once
	ctx := s.withUploadHook(func() error {
		var err error
		once.Do(func() { err = s.renameApp("during-upload") })
		return err
	})
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetTestTimeout(time.Minute)
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))
	env.SetWorkerOptions(worker.Options{BackgroundActivityContext: ctx, DeadlockDetectionTimeout: time.Minute})
	env.RegisterActivityWithOptions(s.activities, activity.RegisterOptions{SkipInvalidStructFunctions: true})
	env.ExecuteWorkflow(statesignals.Regenerate, &pkgstate.ExecuteRegenerationRequest{
		InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		Targets: []pkgstate.PartialTarget{{Name: pkgstate.PartialInputs}},
	})
	require.True(s.T(), env.IsWorkflowCompleted())
	require.NoError(s.T(), env.GetWorkflowError())
	var result pkgstate.ExecuteRegenerationResponse
	require.NoError(s.T(), env.GetWorkflowResult(&result))
	require.Equal(s.T(), "during-upload", result.State.App.Name)
	require.Contains(s.T(), result.UpdatedPartials, pkgstate.PartialApp)

	head := s.head()
	require.True(s.T(), head.StaleAt.Empty())
	stored, fromBlob := s.currentRow().GetState(s.ctx, true)
	require.True(s.T(), fromBlob)
	require.Equal(s.T(), "during-upload", stored.App.Name)
	require.Equal(s.T(), "during-upload", s.fresh().App.Name)
	require.Equal(s.T(), head, s.head(), "a strict read of the published head must not republish")
}
