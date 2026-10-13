package activities_test

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/pkg/types/outputs"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionrunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/syncsecrets"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type producerEnqueue struct {
	OwnerID   string
	OwnerType string
	QueueName string
}

type producerHarness struct {
	s      *InstallStateIntegrationSuite
	env    *testsuite.TestWorkflowEnvironment
	status *statusactivities.Activities

	mu            sync.Mutex
	onRunnerQueue func(ctx context.Context, req *producerEnqueue) error
	hinted        bool
	workerRead    *state.State
	apiRead       *state.State
	readErr       error
}

func (s *InstallStateIntegrationSuite) newProducerHarness() *producerHarness {
	var ws testsuite.WorkflowTestSuite
	h := &producerHarness{
		s:      s,
		env:    ws.NewTestWorkflowEnvironment(),
		status: statusactivities.New(statusactivities.Params{DB: s.deps.DB, MW: s.deps.MW}),
	}
	h.env.SetTestTimeout(time.Minute)
	dc := converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	)
	h.env.SetDataConverter(dc)
	h.env.SetContextPropagators([]workflow.ContextPropagator{propagator.New(propagator.Params{L: zap.NewNop(), DataConverter: dc})})
	h.env.SetWorkerOptions(worker.Options{BackgroundActivityContext: s.ctx, DeadlockDetectionTimeout: time.Minute})
	h.env.RegisterActivityWithOptions(s.activities, activity.RegisterOptions{SkipInvalidStructFunctions: true})

	for name, fn := range map[string]any{
		"UpdateDeployStatusV2":             h.status.UpdateDeployStatusV2,
		"UpdateRunStatusV2":                h.status.UpdateRunStatusV2,
		"UpdateInstallWorkflowRunStatusV2": h.status.UpdateInstallWorkflowRunStatusV2,
		"EnqueueSignalToOwner":             h.enqueueSignalToOwner,
		"CreateLogStream": func(ctx context.Context, req installactivities.CreateLogStreamRequest) (*app.LogStream, error) {
			ls := &app.LogStream{ID: domains.NewLogStreamID(), OrgID: s.install.OrgID, OwnerID: req.StepID, OwnerType: "install_workflow_steps"}
			if req.ActionWorkflowRunID != "" {
				ls.OwnerID = req.ActionWorkflowRunID
				ls.OwnerType = "install_action_workflow_runs"
			}
			return ls, s.deps.DB.WithContext(ctx).Create(ls).Error
		},
		"GetLogStream": func(_ context.Context, req installactivities.GetLogStreamRequest) (*app.LogStream, error) {
			return &app.LogStream{ID: domains.NewLogStreamID(), OrgID: s.install.OrgID, RunnerAPIURL: "http://127.0.0.1:9"}, nil
		},
		"CloseLogStream": func(context.Context, installactivities.CloseLogStreamRequest) error { return nil },
		"GetLatestJob": func(context.Context, *installactivities.GetLatestJobRequest) (*app.RunnerJob, error) {
			return &app.RunnerJob{Execution: &app.RunnerJobExecution{Result: &app.RunnerJobExecutionResult{}}}, nil
		},
		"CreateDeployJob": func(_ context.Context, req *installactivities.CreateDeployJobRequest) (*app.RunnerJob, error) {
			return &app.RunnerJob{ID: domains.NewRunnerJobID(), RunnerID: req.RunnerID, Type: req.Type, Operation: req.Op}, nil
		},
		"CreateSandboxJob": func(_ context.Context, req *installactivities.CreateSandboxJobRequest) (*app.RunnerJob, error) {
			return &app.RunnerJob{ID: domains.NewRunnerJobID(), RunnerID: req.RunnerID, Type: req.JobType, Operation: req.Op}, nil
		},
		"SaveRunnerJobPlan":      func(context.Context, *installactivities.SaveRunnerJobPlanRequest) error { return nil },
		"RecordInstallRoleUsage": func(context.Context, *installactivities.RecordInstallRoleUsageRequest) error { return nil },
	} {
		h.env.RegisterActivityWithOptions(fn, activity.RegisterOptions{Name: name, DisableAlreadyRegisteredCheck: true})
	}

	for name, fn := range map[string]any{
		"CreateDeployPlan": func(workflow.Context, *plan.CreateDeployPlanRequest) (plan.CreateDeployPlanResponse, error) {
			return plan.CreateDeployPlanResponse{Plan: &plantypes.DeployPlan{}}, nil
		},
		"CreateSandboxRunPlan": func(workflow.Context, *plan.CreateSandboxRunPlanRequest) (plan.CreateSandboxPlanResponse, error) {
			return plan.CreateSandboxPlanResponse{Plan: &plantypes.SandboxRunPlan{}}, nil
		},
		"ExecuteJob": func(workflow.Context, *job.ExecuteJobRequest) (app.RunnerJobStatus, error) {
			return app.RunnerJobStatusFinished, nil
		},
	} {
		h.env.RegisterWorkflowWithOptions(fn, workflow.RegisterOptions{Name: name, DisableAlreadyRegisteredCheck: true})
	}
	return h
}

func (h *producerHarness) enqueueSignalToOwner(ctx context.Context, req *producerEnqueue) (*sharedactivities.EnqueueSignalToOwnerResponse, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch req.QueueName {
	case helpers.InstallStateManagerQueueName:
		if !h.hinted {
			h.hinted = true
			h.workerRead, h.readErr = h.s.activities.GetInstallState(ctx, &installactivities.GetInstallStateRequest{InstallID: h.s.install.ID})
			if h.readErr == nil {
				h.apiRead, h.readErr = h.s.deps.Helpers.GetInstallState(ctx, h.s.install.ID, true, false)
			}
		}
	case queuenames.RunnerSignalsQueueName:
		if h.onRunnerQueue != nil {
			if err := h.onRunnerQueue(ctx, req); err != nil {
				return nil, err
			}
		}
	}
	return &sharedactivities.EnqueueSignalToOwnerResponse{QueueSignalID: domains.NewQueueSignalID()}, nil
}

func (h *producerHarness) execute(sig signal.Signal) {
	accountID, err := cctx.AccountIDFromContext(h.s.ctx)
	require.NoError(h.s.T(), err)
	h.env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ctx = cctx.SetAccountIDWorkflowContext(ctx, accountID)
		ctx = cctx.SetOrgIDWorkflowContext(ctx, h.s.install.OrgID)
		return sig.Execute(ctx)
	})
	require.True(h.s.T(), h.env.IsWorkflowCompleted())
	require.ErrorContains(h.s.T(), h.env.GetWorkflowError(), "unable to await state generation")
}

func (h *producerHarness) requirePendingReadsFreshOrStale(what string, fresh func(*state.State) bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	require.True(h.s.T(), h.hinted, "the signal must reach the queued state regeneration: %v", h.env.GetWorkflowError())
	require.NoError(h.s.T(), h.readErr)
	for reader, snapshot := range map[string]*state.State{"worker": h.workerRead, "api": h.apiRead} {
		require.NotNil(h.s.T(), snapshot)
		assert.True(h.s.T(), fresh(snapshot) || snapshot.StaleAt != nil,
			"%s reader returned %s from before the mutation without a stale marker while regeneration was pending", reader, what)
	}
}

func (s *InstallStateIntegrationSuite) saveBaseline(mutate func(*state.State)) {
	current, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	mutate(current)
	helpers.MapLegacyFields(current)
	_, err = s.saveState(current)
	require.NoError(s.T(), err)
}

func (s *InstallStateIntegrationSuite) TestProducerDeployCompletionPublishesComponentState() {
	t := s.T()
	var connection app.ComponentConfigConnection
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Preload("Component").
		Where(app.ComponentConfigConnection{AppConfigID: s.install.AppConfigID, Type: app.ComponentTypeTerraformModule}).
		First(&connection).Error)
	component := connection.Component
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, t, connection.ID)
	installComponent := s.deps.Seeder.CreateInstallComponent(s.ctx, t, s.install.ID, component.ID)
	flow := s.deps.Seeder.CreateWorkflow(s.ctx, t, s.install.ID, app.WorkflowTypeDeployComponents)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, t, flow.ID)
	deploy := s.deps.Seeder.CreateInstallDeploy(s.ctx, t, installComponent.ID, build.ID)
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Model(&app.InstallDeploy{ID: deploy.ID}).
		Updates(app.InstallDeploy{InstallWorkflowID: generics.ToPtr(flow.ID), Status: app.InstallDeployStatusPlanning}).Error)
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.Components = map[string]any{component.Name: map[string]any{
			"populated": true, "name": component.Name, "component_id": component.ID,
			"install_component_id": installComponent.ID, "status": string(app.InstallDeployStatusPlanning),
		}}
	})

	h := s.newProducerHarness()
	h.execute(&componentdeployapplyplan.Signal{
		InstallComponentID: installComponent.ID, InstallID: s.install.ID, ComponentID: component.ID,
		FlowID: flow.ID, FlowStepID: step.ID, InstallWorkflowStepID: step.ID,
	})

	var applied app.InstallDeploy
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Where(app.InstallDeploy{ID: deploy.ID}).First(&applied).Error)
	require.Equal(t, app.InstallDeployStatusActive, applied.Status)
	h.requirePendingReadsFreshOrStale("the component deploy status", func(snapshot *state.State) bool {
		comp, ok := snapshot.Components[component.Name].(map[string]any)
		return ok && comp["status"] == string(app.InstallDeployStatusActive)
	})
}

func (s *InstallStateIntegrationSuite) TestProducerSandboxProvisionPublishesSandboxState() {
	t := s.T()
	flow := s.deps.Seeder.CreateWorkflow(s.ctx, t, s.install.ID, app.WorkflowTypeProvision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, t, flow.ID)
	run := &app.InstallSandboxRun{
		InstallID: s.install.ID, InstallWorkflowID: generics.ToPtr(flow.ID),
		AppSandboxConfigID: s.install.AppSandboxConfigID, RunType: app.SandboxRunTypeProvision,
		Status: app.SandboxRunStatusApplying, StatusDescription: "applying plan",
	}
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(run).Error)
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.Sandbox = &state.SandboxState{Populated: true, Status: string(app.SandboxRunStatusApplying)}
	})

	h := s.newProducerHarness()
	sig := &provisionsandboxapplyplan.Signal{
		InstallSandboxID: s.install.InstallSandbox.ID, InstallID: s.install.ID,
		FlowID: flow.ID, FlowStepID: step.ID, InstallWorkflowID: flow.ID,
	}
	sig.WithParams(&signal.Params{Cfg: s.deps.Cfg})
	h.execute(sig)

	var applied app.InstallSandboxRun
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Where(app.InstallSandboxRun{ID: run.ID}).First(&applied).Error)
	require.Equal(t, app.SandboxRunStatusActive, applied.Status)
	h.requirePendingReadsFreshOrStale("the sandbox run status", func(snapshot *state.State) bool {
		return snapshot.Sandbox != nil && snapshot.Sandbox.Status == string(app.SandboxRunStatusActive)
	})
}

func (s *InstallStateIntegrationSuite) TestProducerRunnerProvisionPublishesRunnerState() {
	t := s.T()
	runner := s.createProducerRunner(app.RunnerStatusProvisioning)
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.Runner = helpers.ToRunnerState(*runner)
	})

	h := s.newProducerHarness()
	h.onRunnerQueue = func(ctx context.Context, req *producerEnqueue) error {
		_, err := h.status.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
			RunnerID: req.OwnerID, Status: app.RunnerStatusAwaitingInstallStackRun, StatusDescription: "awaiting stack run",
		})
		return err
	}
	h.execute(&provisionrunner.Signal{InstallID: s.install.ID})

	var provisioned app.Runner
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Where(app.Runner{ID: runner.ID}).First(&provisioned).Error)
	require.Equal(t, app.RunnerStatusAwaitingInstallStackRun, provisioned.Status)
	h.requirePendingReadsFreshOrStale("the runner status", func(snapshot *state.State) bool {
		return snapshot.Runner != nil && snapshot.Runner.Status == string(app.RunnerStatusAwaitingInstallStackRun)
	})
}

func (s *InstallStateIntegrationSuite) createProducerRunner(status app.RunnerStatus) *app.Runner {
	t := s.T()
	group := &app.RunnerGroup{
		ID: domains.NewRunnerGroupID(), OrgID: s.install.OrgID, OwnerID: s.install.ID, OwnerType: "installs",
		Type: app.RunnerGroupTypeInstall, Platform: app.AppRunnerTypeAWS,
	}
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(group).Error)
	runner := &app.Runner{
		ID: domains.NewRunnerID(), OrgID: s.install.OrgID, Name: "example-runner", DisplayName: "Example Runner",
		Status: status, RunnerGroupID: group.ID,
	}
	require.NoError(t, s.deps.DB.WithContext(s.ctx).Create(runner).Error)
	return runner
}

func (s *InstallStateIntegrationSuite) TestProducerSecretSyncPublishesSecretState() {
	runner := s.createProducerRunner(app.RunnerStatusActive)
	flow := s.deps.Seeder.CreateWorkflow(s.ctx, s.T(), s.install.ID, app.WorkflowTypeProvision)
	step := s.deps.Seeder.CreateWorkflowStep(s.ctx, s.T(), flow.ID)
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.Secrets = generics.ToPtr(state.NewSecretsState())
	})
	h := s.newProducerHarness()
	h.env.RegisterWorkflowWithOptions(func(workflow.Context, *plan.CreateSyncSecretsPlanRequest) (*plantypes.SyncSecretsPlan, error) {
		return &plantypes.SyncSecretsPlan{KubernetesSecrets: []plantypes.KubernetesSecretSync{{SecretName: "example", Namespace: "default", Name: "example", KeyName: "key"}}}, nil
	}, workflow.RegisterOptions{Name: "CreateSyncSecretsPlan"})
	secretOutputs := outputs.SyncSecretsOutput{"example": {Name: "example", Exists: true, KubernetesName: "example", KubernetesKey: "key"}}
	encoded, err := json.Marshal(secretOutputs)
	require.NoError(s.T(), err)
	h.env.RegisterActivityWithOptions(func(ctx context.Context, req *job.ExecuteJobRequest) (app.RunnerJobStatus, error) {
		execution := &app.RunnerJobExecution{RunnerJobID: req.JobID, Status: app.RunnerJobExecutionStatusFinished}
		if err := s.deps.DB.WithContext(ctx).Create(execution).Error; err != nil {
			return "", err
		}
		if err := s.deps.DB.WithContext(ctx).Create(&app.RunnerJobExecutionOutputs{RunnerJobExecutionID: execution.ID, Outputs: encoded}).Error; err != nil {
			return "", err
		}
		return app.RunnerJobStatusFinished, nil
	}, activity.RegisterOptions{Name: "FinishSecretSyncJob"})
	h.env.RegisterWorkflowWithOptions(func(ctx workflow.Context, req *job.ExecuteJobRequest) (app.RunnerJobStatus, error) {
		var status app.RunnerJobStatus
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
		err := workflow.ExecuteActivity(ctx, "FinishSecretSyncJob", req).Get(ctx, &status)
		return status, err
	}, workflow.RegisterOptions{Name: "ExecuteJob", DisableAlreadyRegisteredCheck: true})
	h.execute(&syncsecrets.Signal{InstallID: s.install.ID, WorkflowStepID: step.ID})

	var completed app.RunnerJob
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.RunnerJob{RunnerID: runner.ID, Type: app.RunnerJobTypeSandboxSyncSecrets}).First(&completed).Error)
	require.Equal(s.T(), true, completed.ParsedOutputs["example"].(map[string]any)["exists"])
	h.requirePendingReadsFreshOrStale("the synced secret metadata", func(snapshot *state.State) bool {
		return snapshot.Secrets != nil && (*snapshot.Secrets)["example"].Exists
	})
}

func (s *InstallStateIntegrationSuite) TestProducerActionCompletionPublishesActionState() {
	s.createProducerRunner(app.RunnerStatusActive)
	definition := s.deps.Seeder.CreateActionWorkflow(s.ctx, s.T(), s.install.AppID)
	s.deps.Seeder.CreateActionWorkflowConfig(s.ctx, s.T(), s.install.AppID, s.install.AppConfigID, definition.ID)
	installed := &app.InstallActionWorkflow{InstallID: s.install.ID, ActionWorkflowID: definition.ID}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(installed).Error)
	s.saveBaseline(func(snapshot *state.State) {
		snapshot.Actions = state.NewActionsState()
		snapshot.Actions.Workflows[definition.Name] = &state.ActionWorkflowState{ID: definition.ID, Status: string(app.InstallActionRunStatusUnknown)}
	})
	h := s.newProducerHarness()
	h.env.RegisterWorkflowWithOptions(func(workflow.Context, *plan.CreateActionRunPlanRequest) (plan.CreateActionPlanResponse, error) {
		return plan.CreateActionPlanResponse{Plan: &plantypes.ActionWorkflowRunPlan{}}, nil
	}, workflow.RegisterOptions{Name: "CreateActionWorkflowRunPlan"})
	h.execute(&actionworkflowrun.Signal{
		InstallID: s.install.ID, InstallActionWorkflowID: installed.ID,
		TriggerType: app.ActionWorkflowTriggerTypeManual, TriggeredByID: s.install.ID, TriggeredByType: "installs",
	})
	var completed app.InstallActionWorkflowRun
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallActionWorkflowRun{InstallID: s.install.ID, InstallActionWorkflowID: generics.NewNullString(installed.ID)}).First(&completed).Error)
	require.Equal(s.T(), app.InstallActionRunStatusFinished, completed.Status)
	h.requirePendingReadsFreshOrStale("the action run status", func(snapshot *state.State) bool {
		if snapshot.Actions == nil || snapshot.Actions.Workflows[definition.Name] == nil {
			return false
		}
		return snapshot.Actions.Workflows[definition.Name].Status == string(app.InstallActionRunStatusFinished)
	})
}

func (s *InstallStateIntegrationSuite) TestBackgroundGenerationRebuildsFromCurrentBaseline() {
	older, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "before-upload").Error)
	var once sync.Once
	ctx := blobstore.WithBlobService(s.ctx, publicationBlobService{
		Service: s.deps.Blobs,
		afterUpload: func() error {
			var err error
			once.Do(func() {
				err = s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", "during-upload").Error
			})
			return err
		},
	})
	h := s.newProducerHarness()
	h.env.SetWorkerOptions(worker.Options{BackgroundActivityContext: ctx, DeadlockDetectionTimeout: time.Minute})
	h.env.ExecuteWorkflow(statesignals.Regenerate, &pkgstate.ExecuteRegenerationRequest{
		InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		CachedState: older, Targets: []pkgstate.PartialTarget{{Name: pkgstate.PartialInputs}},
	})
	require.True(s.T(), h.env.IsWorkflowCompleted())
	require.NoError(s.T(), h.env.GetWorkflowError())
	var result pkgstate.ExecuteRegenerationResponse
	require.NoError(s.T(), h.env.GetWorkflowResult(&result))
	require.Equal(s.T(), "during-upload", result.State.App.Name)
	require.Contains(s.T(), result.UpdatedPartials, pkgstate.PartialApp)
	require.NotNil(s.T(), result.State.Revision)
	current, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), result.State.Revision, current.Revision)
	require.Equal(s.T(), "during-upload", current.App.Name)
	require.Nil(s.T(), current.StaleAt)
}
