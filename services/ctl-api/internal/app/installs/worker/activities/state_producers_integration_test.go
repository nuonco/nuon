package activities_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
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
			name: "secrets_sync_outputs",
			commit: func() {
				secretsExec := s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeSandboxSyncSecrets, app.RunnerJobExecutionStatusFinished)
				s.fresh()
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
	for _, legacy := range []bool{false, true} {
		s.Run(fmt.Sprintf("legacy_%t", legacy), func() {
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
			if legacy {
				env.OnGetVersion("install-state-publication", workflow.DefaultVersion, 1).Return(workflow.DefaultVersion)
			}
			env.ExecuteWorkflow(statesignals.Regenerate, &pkgstate.ExecuteRegenerationRequest{
				InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
				Targets: []pkgstate.PartialTarget{{Name: pkgstate.PartialInputs}},
			})
			require.True(s.T(), env.IsWorkflowCompleted())
			require.NoError(s.T(), env.GetWorkflowError())
			var result pkgstate.ExecuteRegenerationResponse
			require.NoError(s.T(), env.GetWorkflowResult(&result))
			if !legacy {
				require.Equal(s.T(), "during-upload", result.State.App.Name)
				require.Contains(s.T(), result.UpdatedPartials, pkgstate.PartialApp)
			}

			head := s.head()
			require.True(s.T(), head.StaleAt.Empty())
			stored, fromBlob := s.currentRow().GetState(s.ctx, true)
			require.True(s.T(), fromBlob)
			require.Equal(s.T(), "during-upload", stored.App.Name)
			require.Equal(s.T(), "during-upload", s.fresh().App.Name)
			require.Equal(s.T(), head, s.head(), "a strict read of the published head must not republish")
		})
	}
}

func (s *InstallStateIntegrationSuite) cleanHead() app.InstallStateHead {
	s.T().Helper()
	s.fresh()
	head := s.head()
	require.True(s.T(), head.StaleAt.Empty())
	require.Empty(s.T(), head.StalePartials)
	return head
}

func (s *InstallStateIntegrationSuite) requireHeadUnchanged(before app.InstallStateHead, regression string) {
	s.T().Helper()
	require.Equal(s.T(), before, s.head(), regression)
}

func (s *InstallStateIntegrationSuite) requireHeadDirty(before app.InstallStateHead, partials ...pkgstate.PartialName) {
	s.T().Helper()
	head := s.head()
	require.Greater(s.T(), head.Revision, before.Revision)
	require.False(s.T(), head.StaleAt.Empty())
	require.Equal(s.T(), partials, head.StalePartials)
}

func (s *InstallStateIntegrationSuite) setColumn(model any, column string, value any) {
	s.T().Helper()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(model).Update(column, value).Error)
}

func secretOutputs(arn string) map[string]any {
	return map[string]any{"example": map[string]any{"name": "example", "arn": arn, "exists": true}}
}

func secretARN(snapshot *state.State) any {
	return (*snapshot.Secrets)["example"].ARN
}

func (s *InstallStateIntegrationSuite) TestRunnerJobNoiseDoesNotInvalidateState() {
	connection := s.componentConnection()
	runner := s.createRunner(app.RunnerStatusActive)
	installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), s.install.ID, connection.ComponentID)
	deploy := s.createDeploy(installed.ID, connection.ID, app.InstallDeployStatusActive)
	deployExec := s.createExecution(runner.ID, deploy.ID, "install_deploys", app.RunnerJobTypeTerraformDeploy, app.RunnerJobExecutionStatusFinished)
	s.create(executionOutputs(deployExec.ID, map[string]any{"endpoint": "api.example.com"}))
	secretsExec := s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeSandboxSyncSecrets, app.RunnerJobExecutionStatusFinished)
	s.create(executionOutputs(secretsExec.ID, secretOutputs("current")))
	requireObserved := func() {
		for range 2 {
			snapshot := s.fresh()
			component := componentState(s.T(), snapshot, connection.Component.Name)
			require.Equal(s.T(), map[string]any{"endpoint": "api.example.com"}, component["outputs"])
			require.Equal(s.T(), "current", secretARN(snapshot))
		}
	}

	for _, tc := range []struct {
		name       string
		regression string
		commit     func()
	}{
		{
			name:       "unrelated_job_on_install_runner",
			regression: "every job on an install runner must not dirty secrets",
			commit: func() {
				healthCheck := s.createExecution(runner.ID, runner.ID, "runners", app.RunnerJobTypeHealthCheck, app.RunnerJobExecutionStatusFinished)
				s.create(executionOutputs(healthCheck.ID, map[string]any{"healthy": true}))
			},
		},
		{
			name:       "runner_job_status_only",
			regression: "runner job status churn must not dirty owner or secrets partials",
			commit: func() {
				for _, execution := range []*app.RunnerJobExecution{deployExec, secretsExec} {
					s.setColumn(&app.RunnerJob{ID: execution.RunnerJobID}, "status", app.RunnerJobStatusInProgress)
					s.setColumn(&app.RunnerJob{ID: execution.RunnerJobID}, "status", app.RunnerJobStatusFinished)
				}
			},
		},
	} {
		s.Run(tc.name, func() {
			before := s.cleanHead()
			requireObserved()
			tc.commit()
			s.requireHeadUnchanged(before, tc.regression)
			requireObserved()
		})
	}

	s.Run("owner_job_on_install_runner_skips_secrets", func() {
		before := s.cleanHead()
		s.createExecution(runner.ID, deploy.ID, "install_deploys", app.RunnerJobTypeTerraformDeploy, app.RunnerJobExecutionStatusFinished)
		s.requireHeadDirty(before, pkgstate.PartialComponents)
		require.Equal(s.T(), "current", secretARN(s.fresh()))
	})
}

func (s *InstallStateIntegrationSuite) TestExecutionTransitionsInvalidateOnlyOutputPriorityChanges() {
	connection := s.componentConnection()
	runner := s.createRunner(app.RunnerStatusActive)
	installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), s.install.ID, connection.ComponentID)
	for _, tc := range []struct {
		name    string
		partial pkgstate.PartialName
		job     func() (string, string, app.RunnerJobType)
		outputs func(string) map[string]any
		observe func(*state.State) any
	}{
		{
			name:    "component_deploy",
			partial: pkgstate.PartialComponents,
			job: func() (string, string, app.RunnerJobType) {
				deploy := s.createDeploy(installed.ID, connection.ID, app.InstallDeployStatusActive)
				return deploy.ID, "install_deploys", app.RunnerJobTypeTerraformDeploy
			},
			outputs: func(endpoint string) map[string]any { return map[string]any{"endpoint": endpoint} },
			observe: func(snapshot *state.State) any {
				outputs, _ := componentState(s.T(), snapshot, connection.Component.Name)["outputs"].(map[string]any)
				return outputs["endpoint"]
			},
		},
		{
			name:    "secrets_sync",
			partial: pkgstate.PartialSecrets,
			job: func() (string, string, app.RunnerJobType) {
				return s.install.ID, "installs", app.RunnerJobTypeSandboxSyncSecrets
			},
			outputs: secretOutputs,
			observe: secretARN,
		},
	} {
		s.Run(tc.name, func() {
			ownerID, ownerType, jobType := tc.job()
			previous := s.createExecution(runner.ID, ownerID, ownerType, jobType, app.RunnerJobExecutionStatusFinished)
			s.create(executionOutputs(previous.ID, tc.outputs("previous")))
			next := &app.RunnerJobExecution{RunnerJobID: previous.RunnerJobID, Status: app.RunnerJobExecutionStatusPending}
			s.create(next)
			s.create(executionOutputs(next.ID, tc.outputs("next")))

			before := s.cleanHead()
			require.Equal(s.T(), "previous", tc.observe(s.fresh()))
			for _, status := range []app.RunnerJobExecutionStatus{
				app.RunnerJobExecutionStatusInProgress,
				app.RunnerJobExecutionStatusFailed,
			} {
				s.setColumn(next, "status", status)
				s.requireHeadUnchanged(before, "non-finished execution transitions must not dirty "+string(tc.partial))
			}

			for _, step := range []struct {
				status app.RunnerJobExecutionStatus
				want   string
			}{
				{status: app.RunnerJobExecutionStatusFinished, want: "next"},
				{status: app.RunnerJobExecutionStatusFailed, want: "previous"},
			} {
				s.setColumn(next, "status", step.status)
				s.requireHeadDirty(before, tc.partial)
				for range 2 {
					require.Equal(s.T(), step.want, tc.observe(s.fresh()))
				}
				before = s.cleanHead()
			}
		})
	}
}

func (s *InstallStateIntegrationSuite) TestRunnerJobTypeChangeInvalidatesOldAndNewSecretsMembership() {
	runner := s.createRunner(app.RunnerStatusActive)
	current := s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeSandboxSyncSecrets, app.RunnerJobExecutionStatusFinished)
	s.create(executionOutputs(current.ID, secretOutputs("current")))
	candidate := s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeHealthCheck, app.RunnerJobExecutionStatusFinished)
	s.create(executionOutputs(candidate.ID, secretOutputs("candidate")))

	before := s.cleanHead()
	require.Equal(s.T(), "current", secretARN(s.fresh()))
	for _, step := range []struct {
		name    string
		jobType app.RunnerJobType
		want    string
	}{
		{name: "joins_secrets_membership", jobType: app.RunnerJobTypeSandboxSyncSecrets, want: "candidate"},
		{name: "leaves_secrets_membership", jobType: app.RunnerJobTypeHealthCheck, want: "current"},
	} {
		s.Run(step.name, func() {
			s.setColumn(&app.RunnerJob{ID: candidate.RunnerJobID}, "type", step.jobType)
			s.requireHeadDirty(before, pkgstate.PartialSecrets)
			for range 2 {
				require.Equal(s.T(), step.want, secretARN(s.fresh()))
			}
			before = s.cleanHead()
		})
	}
}

func (s *InstallStateIntegrationSuite) TestSecretsTypeChangeWaitsForEarlierUnmatchedOutputs() {
	runner := s.createRunner(app.RunnerStatusActive)
	current := s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeSandboxSyncSecrets, app.RunnerJobExecutionStatusFinished)
	s.create(executionOutputs(current.ID, secretOutputs("current")))
	candidate := s.createExecution(runner.ID, s.install.ID, "installs", app.RunnerJobTypeHealthCheck, app.RunnerJobExecutionStatusFinished)
	before := s.cleanHead()

	tx := s.deps.DB.WithContext(s.ctx).Begin()
	require.NoError(s.T(), tx.Error)
	defer tx.Rollback()
	require.NoError(s.T(), tx.Create(executionOutputs(candidate.ID, secretOutputs("candidate"))).Error)
	s.requireHeadUnchanged(before, "the output transaction still sees a non-secrets job")
	s.setColumn(&app.RunnerJob{ID: candidate.RunnerJobID}, "type", app.RunnerJobTypeSandboxSyncSecrets)
	s.requireHeadDirty(before, pkgstate.PartialSecrets)

	attempted := s.observeSourceBarrier()
	done := make(chan struct{})
	var snapshot *state.State
	var readErr error
	go func() {
		defer close(done)
		snapshot, readErr = s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, false, true)
	}()
	select {
	case <-attempted:
	case <-done:
		s.FailNow("state must wait for outputs whose eligibility changed in another transaction", "%v", readErr)
	case <-time.After(30 * time.Second):
		s.FailNow("timed out waiting for the source barrier")
	}
	require.NoError(s.T(), tx.Commit().Error)
	s.await(done)
	require.NoError(s.T(), readErr)
	require.Equal(s.T(), "candidate", secretARN(snapshot))
	require.Equal(s.T(), "candidate", secretARN(s.fresh()))
}

func (s *InstallStateIntegrationSuite) TestActionRunStatusChangesReachStrictFreshReads() {
	definition := s.deps.Seeder.CreateActionWorkflow(s.ctx, s.T(), s.install.AppID)
	s.deps.Seeder.CreateActionWorkflowConfig(s.ctx, s.T(), s.install.AppID, s.install.AppConfigID, definition.ID)
	installed := &app.InstallActionWorkflow{InstallID: s.install.ID, ActionWorkflowID: definition.ID}
	s.create(installed)
	run := &app.InstallActionWorkflowRun{
		InstallID: s.install.ID, InstallActionWorkflowID: generics.NewNullString(installed.ID),
		Status: app.InstallActionRunStatusQueued, StatusDescription: "queued",
		TriggerType: app.ActionWorkflowTriggerTypeManual,
	}
	s.create(run)

	before := s.cleanHead()
	require.Equal(s.T(), string(app.InstallActionRunStatusQueued), s.fresh().Actions.Workflows[definition.Name].Status)
	for _, status := range []app.InstallActionWorkflowRunStatus{
		app.InstallActionRunStatusInProgress,
		app.InstallActionRunStatusError,
		app.InstallActionRunStatusFinished,
	} {
		s.Run(string(status), func() {
			s.setColumn(run, "status", status)
			s.requireHeadDirty(before, pkgstate.PartialActions)
			for range 2 {
				require.Equal(s.T(), string(status), s.fresh().Actions.Workflows[definition.Name].Status)
			}
			before = s.cleanHead()
		})
	}
}
