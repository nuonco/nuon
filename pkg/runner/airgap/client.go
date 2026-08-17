package airgap

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

var _ nuonrunner.Client = (*Client)(nil)

type Result struct {
	Succeeded  bool
	FailedStep string
}

type ClientOptions struct {
	BundleDigest      string
	BundleUpgrade     bool
	UpgradeSandbox    bool
	UpgradeComponents []string
	RequireApproval   bool
	ApprovalPhase     string
}

type Client struct {
	envelope            *Envelope
	store               statestore.Store
	logger              *zap.Logger
	mu                  sync.Mutex
	status              *statestore.Status
	executions          map[string]*models.AppRunnerJobExecution
	results             map[string]*models.ServiceCreateRunnerJobExecutionResultRequest
	process             *models.AppRunnerProcess
	installStackOutputs map[string]any
	installInputs       map[string]string
	done                chan struct{}
	doneOnce            sync.Once
	result              Result
	runtimeJobs         map[string]*RuntimeJob
	runtimeQueue        []string
	applyApproved       bool
	upgradeSandbox      bool
	upgradeComponents   int

	healthMu     sync.Mutex
	healthLoaded bool
	lastHealth   *HealthSnapshot

	refSnapshotsOnce sync.Once
	refInstallStack  map[string]any
	refSandbox       map[string]any
}

// SetInstallStackOutputs supplies the target environment's install stack
// outputs so step plans are rebound away from the reference install's stale
// rendered values. Call before the runner starts executing steps.
//
// Non-string values are normalized to their JSON encoding. The control plane
// stores phone-home payloads in an hstore column (flat string values), and the
// exported plans were rendered from that representation — so a raw phone-home
// payload (e.g. the air-gapped S3 rendezvous object, which nests objects like
// install_inputs) must be flattened the same way for value matching to work.
func (c *Client) SetInstallStackOutputs(outputs map[string]any) {
	normalized := make(map[string]any, len(outputs))
	for key, value := range outputs {
		if _, ok := value.(string); ok {
			normalized[key] = value
			continue
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			continue
		}
		normalized[key] = string(encoded)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.installStackOutputs = normalized
}

// SetInstallInputs supplies customer values for the envelope's bindable
// install inputs. Call before the runner starts executing steps; values are
// substituted for input placeholders in every rendered step plan.
func (c *Client) SetInstallInputs(values map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.installInputs = values
}

func NewClient(envelope *Envelope, store statestore.Store, logger *zap.Logger) (*Client, error) {
	return NewClientWithOptions(envelope, store, logger, ClientOptions{})
}

func NewClientWithOptions(envelope *Envelope, store statestore.Store, logger *zap.Logger, options ClientOptions) (*Client, error) {
	now := time.Now().UTC()
	status, err := store.ReadStatus()
	if err != nil {
		return nil, fmt.Errorf("read airgap status: %w", err)
	}
	if status != nil && status.InstallID != "" && status.InstallID != envelope.InstallID {
		// A different deployment ID would make plan resource names diverge from persisted state.
		return nil, fmt.Errorf("state was initialized for install %q but this envelope resolves to %q; resume with the same --deployment-id the run was started with", status.InstallID, envelope.InstallID)
	}
	if status == nil {
		status = &statestore.Status{InstallID: envelope.InstallID, BundleDigest: options.BundleDigest, RunID: uuid.NewString(), RunType: statestore.RunTypeInstall, Status: statestore.RunStatusInProgress, StartedAt: now, HeartbeatAt: now, Outputs: map[string]json.RawMessage{}}
		for _, step := range envelope.Steps {
			status.Steps = append(status.Steps, statestore.StepStatus{ID: step.ID, Name: step.Name, Status: string(models.AppRunnerJobStatusAvailable)})
		}
		if err := store.WriteStatus(status); err != nil {
			return nil, fmt.Errorf("initialize airgap status: %w", err)
		}
	} else if options.BundleDigest != "" && status.BundleDigest != "" && status.BundleDigest != options.BundleDigest {
		if !options.BundleUpgrade {
			return nil, fmt.Errorf("bundle digest changed from %q to %q without a staged upgrade candidate", status.BundleDigest, options.BundleDigest)
		}
		changed := make(map[string]bool, len(options.UpgradeComponents))
		for _, component := range options.UpgradeComponents {
			changed[component] = true
		}
		if err := store.WriteStatus(status); err != nil {
			return nil, fmt.Errorf("archive previous airgap run: %w", err)
		}
		previousRunID := status.RunID
		previousSteps := make(map[string]statestore.StepStatus, len(status.Steps))
		for _, step := range status.Steps {
			previousSteps[step.ID] = step
		}
		status = &statestore.Status{
			InstallID: envelope.InstallID, BundleDigest: options.BundleDigest, RunID: uuid.NewString(),
			RunType: statestore.RunTypeUpgrade, PreviousRunID: previousRunID,
			Status: statestore.RunStatusInProgress, StartedAt: now, HeartbeatAt: now,
			Outputs: status.Outputs, ApprovalRequired: options.RequireApproval, ApprovalPhase: options.ApprovalPhase,
		}
		if options.RequireApproval {
			status.ResultDirective = statestore.DirectiveAwaitApproval
		}
		for _, step := range envelope.Steps {
			stepStatus := string(models.AppRunnerJobStatusAvailable)
			sourceRunID := ""
			var resultDirective statestore.ResultDirective
			previousStep, found := previousSteps[step.ID]
			if found && previousStep.Status == string(models.AppRunnerJobStatusFinished) {
				stepStatus = string(models.AppStatusAutoDashSkipped)
				sourceRunID = previousRunID
				resultDirective = statestore.DirectiveSkipGroup
			}
			if options.UpgradeSandbox && (step.ID == "sandbox-plan" || step.ID == "sandbox-apply") {
				stepStatus = string(models.AppRunnerJobStatusAvailable)
				sourceRunID = ""
				resultDirective = ""
			}
			for component := range changed {
				if step.ID == "sync-"+component || step.ID == "deploy-"+component+"-plan" || step.ID == "deploy-"+component+"-apply" {
					stepStatus = string(models.AppRunnerJobStatusAvailable)
					sourceRunID = ""
					resultDirective = ""
					break
				}
			}
			status.Steps = append(status.Steps, statestore.StepStatus{ID: step.ID, Name: step.Name, Status: stepStatus, SourceRunID: sourceRunID, ResultDirective: resultDirective})
		}
		if err := store.WriteStatus(status); err != nil {
			return nil, fmt.Errorf("initialize bundle upgrade status: %w", err)
		}
	} else {
		if status.BundleDigest == "" {
			status.BundleDigest = options.BundleDigest
		}
		if status.Outputs == nil {
			status.Outputs = map[string]json.RawMessage{}
		}
		if status.Status == statestore.RunStatusInProgress {
			status.Status = statestore.RunStatusFailed
			status.ResultDirective = statestore.DirectiveRetryGroup
			now := time.Now().UTC()
			status.FinishedAt = &now
			for i := range status.Steps {
				if status.Steps[i].Status == string(models.AppRunnerJobStatusInDashProgress) {
					status.Steps[i].Status = string(models.AppRunnerJobStatusCancelled)
					status.Steps[i].FinishedAt = &now
				}
			}
			if err := store.WriteStatus(status); err != nil {
				return nil, fmt.Errorf("terminalize interrupted airgap run: %w", err)
			}
			status = retryStatus(envelope, status, now)
			if err := store.WriteStatus(status); err != nil {
				return nil, fmt.Errorf("initialize retry after interruption: %w", err)
			}
		} else if status.Status != statestore.RunStatusFailedPendingRetry {
			if err := store.WriteStatus(status); err != nil {
				return nil, fmt.Errorf("persist airgap status: %w", err)
			}
		}
	}
	c := &Client{envelope: envelope, store: store, logger: logger, status: status, executions: map[string]*models.AppRunnerJobExecution{}, results: map[string]*models.ServiceCreateRunnerJobExecutionResultRequest{}, runtimeJobs: map[string]*RuntimeJob{}, done: make(chan struct{}), applyApproved: !status.ApprovalRequired, upgradeSandbox: options.UpgradeSandbox, upgradeComponents: len(options.UpgradeComponents)}
	c.checkDoneLocked()
	return c, nil
}

func retryStatus(envelope *Envelope, previous *statestore.Status, now time.Time) *statestore.Status {
	replan := map[string]bool{}
	for _, step := range envelope.Steps {
		if step.PlanFromStep != "" {
			prior := findStepStatus(previous.Steps, step.ID)
			if prior == nil || prior.Status != string(models.AppRunnerJobStatusFinished) {
				replan[step.PlanFromStep] = true
			}
		}
	}
	next := &statestore.Status{InstallID: previous.InstallID, BundleDigest: previous.BundleDigest, RunID: uuid.NewString(), RunType: previous.RunType, PreviousRunID: previous.RunID, Status: statestore.RunStatusInProgress, StartedAt: now, HeartbeatAt: now, Outputs: previous.Outputs, ApprovalRequired: previous.ApprovalRequired, ApprovalPhase: previous.ApprovalPhase}
	for _, step := range envelope.Steps {
		prior := findStepStatus(previous.Steps, step.ID)
		status, source := string(models.AppRunnerJobStatusAvailable), ""
		var resultDirective statestore.ResultDirective
		if prior != nil && prior.Status == string(models.AppRunnerJobStatusFinished) && !replan[step.ID] {
			status, source = string(models.AppStatusAutoDashSkipped), previous.RunID
			resultDirective = statestore.DirectiveSkipGroup
		}
		next.Steps = append(next.Steps, statestore.StepStatus{ID: step.ID, Name: step.Name, Status: status, SourceRunID: source, ResultDirective: resultDirective})
	}
	return next
}

func findStepStatus(steps []statestore.StepStatus, id string) *statestore.StepStatus {
	for i := range steps {
		if steps[i].ID == id {
			return &steps[i]
		}
	}
	return nil
}

func (c *Client) Done() <-chan struct{} { return c.done }
func (c *Client) Result() Result {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.result
}

func (c *Client) ApproveApply() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.applyApproved = true
	c.status.ApprovalRequired = false
	c.status.ResultDirective = statestore.DirectiveContinue
	if err := c.store.WriteStatus(c.status); err != nil {
		return err
	}
	c.checkDoneLocked()
	return nil
}

func (c *Client) Approval() (bool, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status.ApprovalRequired, c.status.ApprovalPhase
}
func (c *Client) SetRunnerID(string)  {}
func (c *Client) SetAuthToken(string) {}

func (c *Client) GetSettings(context.Context) (*models.AppRunnerGroupSettings, error) {
	return &models.AppRunnerGroupSettings{Groups: []string{"sandbox", "deploy", "sync", "actions"}, HeartBeatTimeout: int64(30 * time.Second), LoggingLevel: "info", Metadata: map[string]string{}, EnableLogging: false, EnableMetrics: false, LongPollJobs: false}, nil
}

func (c *Client) GetJobs(_ context.Context, group models.AppRunnerJobGroup, _ models.AppRunnerJobStatus, _ *int64) ([]*models.AppRunnerJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, status := range c.status.Steps {
		if status.Status == string(models.AppRunnerJobStatusInDashProgress) {
			return nil, nil
		}
	}
	if c.result.FailedStep != "" {
		return nil, nil
	}
	for _, step := range c.envelope.Steps {
		if models.AppRunnerJobGroup(step.JobGroup) != group || c.stepStatus(step.ID).Status != string(models.AppRunnerJobStatusAvailable) || !c.dependenciesFinished(step) || (step.JobOperation == "apply-plan" && !c.applyApproved) {
			continue
		}
		return []*models.AppRunnerJob{c.job(step)}, nil
	}
	if c.result.Succeeded && len(c.runtimeQueue) > 0 {
		runtime := c.runtimeJobs[c.runtimeQueue[0]]
		if runtime != nil && runtime.JobGroup == string(group) && !runtime.started {
			return []*models.AppRunnerJob{c.runtimeJob(runtime)}, nil
		}
	}
	return nil, nil
}

func (c *Client) TailJobs(ctx context.Context, group models.AppRunnerJobGroup, wait time.Duration) ([]*models.AppRunnerJob, error) {
	if wait > 2*time.Second {
		wait = 2 * time.Second
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(wait):
		return c.GetJobs(ctx, group, models.AppRunnerJobStatusAvailable, nil)
	}
}

func (c *Client) GetJob(_ context.Context, id string) (*models.AppRunnerJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if runtime := c.runtimeJobs[id]; runtime != nil {
		return c.runtimeJob(runtime), nil
	}
	step, err := c.findStep(id)
	if err != nil {
		return nil, err
	}
	return c.job(*step), nil
}

func (c *Client) GetJobPlanJSON(_ context.Context, id string) (string, error) {
	if runtime := c.getRuntimeJob(id); runtime != nil {
		plan, err := c.renderPlanWithEnvelope(id, runtime.CompositePlan, "", runtime.Envelope)
		return string(plan), err
	}
	step, err := c.findStep(id)
	if err != nil {
		return "", err
	}
	plan, err := c.renderStepPlan(step)
	if err != nil {
		return "", err
	}
	c.persistStepPlan(step.ID, plan)
	return string(plan), nil
}

func (c *Client) GetJobCompositePlan(_ context.Context, id string) (*models.PlantypesCompositePlan, error) {
	if runtime := c.getRuntimeJob(id); runtime != nil {
		raw, err := c.renderPlanWithEnvelope(id, runtime.CompositePlan, "", runtime.Envelope)
		if err != nil {
			return nil, err
		}
		var plan models.PlantypesCompositePlan
		if err := json.Unmarshal(raw, &plan); err != nil {
			return nil, fmt.Errorf("decode composite plan for %s: %w", id, err)
		}
		return &plan, nil
	}
	step, err := c.findStep(id)
	if err != nil {
		return nil, err
	}
	raw, err := c.renderStepPlan(step)
	if err != nil {
		return nil, err
	}
	var plan models.PlantypesCompositePlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		return nil, fmt.Errorf("decode composite plan for %s: %w", id, err)
	}
	c.persistStepPlan(step.ID, raw)
	return &plan, nil
}

// persistStepPlan mirrors the late-bound rendered plan into the state store so
// the portal can preview what a bootstrap step is about to apply. The write is
// best-effort: rendering is authoritative and must not fail with it.
func (c *Client) persistStepPlan(id string, plan json.RawMessage) {
	if err := c.store.WriteFile(statestore.StepPlanKey(id), plan); err != nil {
		c.logger.Warn("persist rendered step plan", zap.String("step_id", id), zap.Error(err))
	}
}

func (c *Client) UpdateJob(ctx context.Context, id string, req *models.ServiceUpdateRunnerJobRequest) (*models.AppRunnerJob, error) {
	return c.GetJob(ctx, id)
}

func (c *Client) GetJobExecutions(_ context.Context, id string) ([]*models.AppRunnerJobExecution, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if execution := c.executions[id]; execution != nil {
		return []*models.AppRunnerJobExecution{execution}, nil
	}
	return nil, nil
}

func (c *Client) CreateJobExecution(_ context.Context, id string, _ *models.ServiceCreateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if runtime := c.runtimeJobs[id]; runtime != nil {
		now := time.Now().UTC()
		execution := &models.AppRunnerJobExecution{ID: "airgap-" + uuid.NewString()[:8], RunnerJobID: id, Status: models.AppRunnerJobExecutionStatusPending, CreatedAt: now.Format(time.RFC3339Nano)}
		c.executions[id], runtime.started = execution, true
		if err := c.writeRuntimeJSON(runtime, "executions.json", []*models.AppRunnerJobExecution{execution}); err != nil {
			return nil, err
		}
		return execution, nil
	}
	status := c.stepStatus(id)
	if status == nil {
		return nil, fmt.Errorf("unknown airgap step %q", id)
	}
	now := time.Now().UTC()
	execution := &models.AppRunnerJobExecution{ID: "airgap-" + uuid.NewString()[:8], RunnerJobID: id, Status: models.AppRunnerJobExecutionStatusPending, CreatedAt: now.Format(time.RFC3339Nano)}
	c.executions[id] = execution
	status.Status = string(models.AppRunnerJobStatusInDashProgress)
	status.ExecutionID = execution.ID
	status.StartedAt = &now
	c.status.HeartbeatAt = now
	if err := c.store.AppendExecution(id, execution); err != nil {
		return nil, err
	}
	if err := c.store.WriteStatus(c.status); err != nil {
		return nil, err
	}
	return execution, nil
}

func (c *Client) UpdateJobExecution(_ context.Context, id, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	execution := c.executions[id]
	if execution == nil || execution.ID != executionID {
		return nil, fmt.Errorf("unknown execution %q", executionID)
	}
	execution.Status = req.Status
	if runtime := c.runtimeJobs[id]; runtime != nil {
		if err := c.writeRuntimeJSON(runtime, "executions.json", []*models.AppRunnerJobExecution{execution}); err != nil {
			return nil, err
		}
		switch req.Status {
		case models.AppRunnerJobExecutionStatusFinished:
			c.completeRuntimeLocked(runtime, true, "")
		case models.AppRunnerJobExecutionStatusFailed, models.AppRunnerJobExecutionStatusCancelled, models.AppRunnerJobExecutionStatusTimedDashOut:
			c.completeRuntimeLocked(runtime, false, req.StatusDescription)
		}
		return execution, nil
	}
	status := c.stepStatus(id)
	now := time.Now().UTC()
	c.status.HeartbeatAt = now
	switch req.Status {
	case models.AppRunnerJobExecutionStatusFinished:
		status.Status = string(models.AppRunnerJobStatusFinished)
		status.FinishedAt = &now
		status.ResultDirective = statestore.DirectiveContinue
		if id == "sandbox-apply" && c.upgradeSandbox && c.status.ApprovalPhase == "sandbox" {
			if c.upgradeComponents > 0 {
				c.status.ApprovalRequired = true
				c.status.ApprovalPhase = "components"
				c.applyApproved = false
			} else {
				c.status.ApprovalPhase = ""
			}
		}
	case models.AppRunnerJobExecutionStatusFailed, models.AppRunnerJobExecutionStatusCancelled, models.AppRunnerJobExecutionStatusTimedDashOut:
		status.Status = string(req.Status)
		status.Error = req.StatusDescription
		status.FinishedAt = &now
		status.ResultDirective = statestore.DirectiveAwaitRetry
		c.result = Result{FailedStep: id}
	default:
		status.Status = string(models.AppRunnerJobStatusInDashProgress)
	}
	if err := c.store.AppendExecution(id, execution); err != nil {
		return nil, err
	}
	if err := c.store.WriteStatus(c.status); err != nil {
		return nil, err
	}
	c.checkDoneLocked()
	return execution, nil
}

func (c *Client) CreateJobExecutionResult(_ context.Context, id, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	if runtime := c.getRuntimeJob(id); runtime != nil {
		if err := c.writeRuntimeJSON(runtime, "result.json", req); err != nil {
			return nil, err
		}
		c.mu.Lock()
		runtime.result = req
		c.mu.Unlock()
		return &models.AppRunnerJobExecutionResult{ID: "result-" + executionID, RunnerJobExecutionID: executionID, Contents: req.Contents, Success: req.Success, ErrorCode: req.ErrorCode, ErrorMetadata: req.ErrorMetadata}, nil
	}
	if err := c.store.WriteResult(id, req); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.results[id] = req
	c.mu.Unlock()
	return &models.AppRunnerJobExecutionResult{ID: "result-" + executionID, RunnerJobExecutionID: executionID, Contents: req.Contents, Success: req.Success, ErrorCode: req.ErrorCode, ErrorMetadata: req.ErrorMetadata}, nil
}

func (c *Client) CreateJobExecutionOutputs(_ context.Context, id, executionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error) {
	b, err := json.Marshal(req.Outputs)
	if err != nil {
		return nil, err
	}
	if runtime := c.getRuntimeJob(id); runtime != nil {
		if err := c.writeRuntimeJSON(runtime, "outputs.json", req); err != nil {
			return nil, err
		}
		c.mu.Lock()
		runtime.outputs = req
		c.mu.Unlock()
		var outputs map[string]any
		_ = json.Unmarshal(b, &outputs)
		return &models.AppRunnerJobExecutionOutputs{ID: "outputs-" + executionID, RunnerJobExecutionID: executionID, Outputs: outputs, OutputsJSON: string(b)}, nil
	}
	if err := c.store.WriteOutputs(id, req); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.Outputs[id] = b
	if err := c.store.WriteStatus(c.status); err != nil {
		return nil, err
	}
	var outputs map[string]any
	_ = json.Unmarshal(b, &outputs)
	return &models.AppRunnerJobExecutionOutputs{ID: "outputs-" + executionID, RunnerJobExecutionID: executionID, Outputs: outputs, OutputsJSON: string(b)}, nil
}

// UpdateTerraformStateJSON receives the `terraform show -json` document the
// handlers upload after an apply. ctl-api stores this in a dedicated column,
// separate from the http-backend state, so it must never overwrite the state
// that terraform reads back through the loopback backend.
func (c *Client) UpdateTerraformStateJSON(_ context.Context, workspaceID string, _ *string, body any) (any, error) {
	var b []byte
	switch v := body.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	return body, c.store.PutTFStateShow(workspaceID, b)
}
func (c *Client) LockTerraformWorkspace(_ context.Context, workspaceID string, _ *string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.store.LockTF(workspaceID, b)
}
func (c *Client) UnlockTerraformWorkspace(_ context.Context, workspaceID string) error {
	return c.store.UnlockTF(workspaceID)
}

func (c *Client) GetAppConfig(_ context.Context, _, _ string) (*models.AppAppConfig, error) {
	if len(c.envelope.AppConfig) == 0 {
		return nil, fmt.Errorf("airgap envelope has no app config")
	}
	var config models.AppAppConfig
	if err := json.Unmarshal(c.envelope.AppConfig, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Client) CreateHeartBeat(context.Context, *models.ServiceCreateRunnerHeartBeatRequest) (*models.AppRunnerHeartBeat, error) {
	return &models.AppRunnerHeartBeat{ID: "airgap-heartbeat"}, nil
}
func (c *Client) CreateHealthCheck(context.Context, *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error) {
	return &models.AppRunnerHealthCheck{ID: "airgap-healthcheck"}, nil
}

// CreateComponentHealth persists the report to the state store instead of
// POSTing it to ctl-api: the latest snapshot is overwritten and any aggregate
// health changes are appended as immutable transitions.
func (c *Client) CreateComponentHealth(_ context.Context, req *models.ServiceCreateComponentHealthRequest) (*models.ServiceCreateComponentHealthResponse, error) {
	snapshot := NewHealthSnapshot(req, c.envelope.Components)

	c.healthMu.Lock()
	defer c.healthMu.Unlock()
	if !c.healthLoaded {
		// Seed the previous snapshot from disk so a restarted runner does not
		// record spurious transitions for components whose health is unchanged.
		if raw, ok, err := c.store.ReadHealth(); err == nil && ok {
			var previous HealthSnapshot
			if err := json.Unmarshal(raw, &previous); err == nil {
				c.lastHealth = &previous
			}
		}
		c.healthLoaded = true
	}

	transitions := HealthTransitions(c.lastHealth, snapshot)
	if err := c.store.WriteHealth(snapshot); err != nil {
		return nil, err
	}
	if len(transitions) > 0 {
		values := make([]any, 0, len(transitions))
		for _, t := range transitions {
			values = append(values, t)
		}
		if err := c.store.AppendHealthTransitions(values); err != nil {
			return nil, err
		}
		for _, t := range transitions {
			c.logger.Info("component health transition",
				zap.String("component", t.ComponentName),
				zap.String("from", t.From),
				zap.String("to", t.To))
		}
	}
	c.lastHealth = snapshot
	return &models.ServiceCreateComponentHealthResponse{}, nil
}

func (c *Client) LatestHealth() *HealthSnapshot {
	c.healthMu.Lock()
	defer c.healthMu.Unlock()
	if c.lastHealth == nil {
		return nil
	}
	copy := *c.lastHealth
	return &copy
}

func (c *Client) GetRunnerInstallComponents(context.Context) (*models.ServiceRunnerInstallComponentsResponse, error) {
	components := make([]*models.ServiceRunnerInstallComponent, 0, len(c.envelope.Components))
	for _, spec := range c.envelope.Components {
		components = append(components, &models.ServiceRunnerInstallComponent{
			InstallComponentID: spec.InstallComponentID,
			ComponentID:        spec.ComponentID,
			ComponentName:      spec.ComponentName,
			ComponentType:      spec.ComponentType,
			HelmReleaseName:    spec.HelmReleaseName,
			HelmNamespace:      spec.HelmNamespace,
		})
	}
	return &models.ServiceRunnerInstallComponentsResponse{InstallID: c.envelope.InstallID, Components: components}, nil
}

type healthContext struct {
	ClusterInfoJSON string   `json:"cluster_info_json,omitempty"`
	SandboxReleases []string `json:"sandbox_releases,omitempty"`
	ComponentKinds  []string `json:"component_kinds,omitempty"`
}

func (c *Client) PutComponentHealthContext(_ context.Context, clusterInfoJSON string, releases, kinds []string) error {
	return c.store.WriteHealthContext(healthContext{ClusterInfoJSON: clusterInfoJSON, SandboxReleases: releases, ComponentKinds: kinds})
}
func (c *Client) GetComponentHealthContext(context.Context) (string, []string, []string, error) {
	raw, ok, err := c.store.ReadHealthContext()
	if err != nil || !ok {
		return "", nil, nil, err
	}
	var stored healthContext
	if err := json.Unmarshal(raw, &stored); err != nil {
		return "", nil, nil, fmt.Errorf("decode health context: %w", err)
	}
	return stored.ClusterInfoJSON, stored.SandboxReleases, stored.ComponentKinds, nil
}

func (c *Client) CreateProcess(_ context.Context, _ *models.ServiceCreateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.process == nil {
		c.process = &models.AppRunnerProcess{ID: "airgap-process", RunnerID: "airgap-" + c.envelope.InstallID, LogStreamID: ""}
	}
	return c.process, nil
}
func (c *Client) GetProcess(context.Context, string) (*models.AppRunnerProcess, error) {
	return c.process, nil
}
func (c *Client) GetProcessShutdowns(context.Context, string) ([]*models.AppRunnerProcessShutdown, error) {
	return nil, nil
}
func (c *Client) UpdateProcess(context.Context, string, *models.ServiceUpdateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	return c.process, nil
}
func (c *Client) CompleteShutdown(context.Context, string, string) (*models.AppRunnerProcessShutdown, error) {
	return &models.AppRunnerProcessShutdown{}, nil
}
func (c *Client) ReportTerminating(context.Context, string) error { return nil }
func (c *Client) GetRunner(context.Context) (*models.AppRunner, error) {
	return &models.AppRunner{ID: "airgap-" + c.envelope.InstallID, OrgID: c.envelope.OrgID, Status: "active"}, nil
}
func (c *Client) GetSandboxConfigs(context.Context) ([]*nuonrunner.SandboxConfig, error) {
	return nil, nil
}
func (c *Client) GetSandboxConfig(context.Context, string, string) (*nuonrunner.SandboxConfig, error) {
	return nil, nil
}
func (c *Client) GetInstallComponenetLastActivePlan(context.Context, string, string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error) {
	return &models.ServiceGetInstallComponenetLastActivePlanResponse{}, nil
}

func (c *Client) WriteOTELLogs(context.Context, interface{}) error    { return nil }
func (c *Client) WriteOTELTraces(context.Context, interface{}) error  { return nil }
func (c *Client) WriteOTELMetrics(context.Context, interface{}) error { return nil }

func unsupported(method string) error { return fmt.Errorf("%s is not supported in airgap M0", method) }
func (c *Client) UpdateInstallActionWorkflowRunStep(_ context.Context, _ string, workflowID, stepID string, req *models.ServiceUpdateInstallActionWorkflowRunStepRequest) (*models.AppInstallActionWorkflowRunStep, error) {
	job := c.runtimeActionJob(workflowID)
	if job == nil {
		return nil, unsupported("UpdateInstallActionWorkflowRunStep")
	}
	_ = c.writeRuntimeJSON(job, "action-step-"+stepID+".json", req)
	return &models.AppInstallActionWorkflowRunStep{ID: stepID, StepID: stepID, InstallActionWorkflowRunID: workflowID}, nil
}
func (c *Client) GetInstallActionWorkflowRun(ctx context.Context, _ string, runID string) (*models.AppInstallActionWorkflowRun, error) {
	job := c.runtimeActionJob(runID)
	if job == nil {
		return nil, unsupported("GetInstallActionWorkflowRun")
	}
	plan, err := c.GetJobCompositePlan(ctx, job.RuntimeJobID)
	if err != nil || plan.ActionWorkflowRunPlan == nil {
		return nil, fmt.Errorf("decode day-2 action workflow run %q: %w", runID, err)
	}
	action := plan.ActionWorkflowRunPlan
	run := &models.AppInstallActionWorkflowRun{ID: action.ID, InstallID: action.InstallID, ActionWorkflowConfigID: "", RunEnvVars: action.OverrideEnvVars, Timeout: action.Timeout}
	for i, step := range action.Steps {
		name := step.Attrs["name"]
		if name == "" {
			name = fmt.Sprintf("step-%d", i+1)
		}
		run.Steps = append(run.Steps, &models.AppInstallActionWorkflowRunStep{ID: step.RunID, StepID: step.RunID, InstallActionWorkflowRunID: action.ID, AdhocConfig: &models.AppAdHocStepConfig{Name: name, Command: step.InterpolatedCommand, InlineContents: step.InterpolatedInlineContents, EnvVars: step.InterpolatedEnvVars, Idx: int64(i)}})
	}
	return run, nil
}
func (c *Client) GetActionWorkflowConfig(context.Context, string) (*models.AppActionWorkflowConfig, error) {
	return nil, unsupported("GetActionWorkflowConfig")
}
func (c *Client) RunnerAuthAWS(context.Context, *models.ServiceRunnerAuthAWSRequest) (*models.ServiceRunnerAuthAWSResponse, error) {
	return nil, unsupported("RunnerAuthAWS")
}
func (c *Client) RunnerAuthAWSIID(context.Context, *models.ServiceRunnerAuthAWSIIDRequest) (*models.ServiceRunnerAuthAWSIIDResponse, error) {
	return nil, unsupported("RunnerAuthAWSIID")
}
func (c *Client) RunnerAuthGCP(context.Context, *models.ServiceRunnerAuthGCPRequest) (*models.ServiceRunnerAuthGCPResponse, error) {
	return nil, unsupported("RunnerAuthGCP")
}
func (c *Client) RunnerAuthAzure(context.Context, *models.ServiceRunnerAuthAzureRequest) (*models.ServiceRunnerAuthAzureResponse, error) {
	return nil, unsupported("RunnerAuthAzure")
}

func (c *Client) findStep(id string) (*Step, error) {
	for i := range c.envelope.Steps {
		if c.envelope.Steps[i].ID == id {
			return &c.envelope.Steps[i], nil
		}
	}
	return nil, fmt.Errorf("unknown airgap step %q", id)
}
func (c *Client) stepStatus(id string) *statestore.StepStatus {
	for i := range c.status.Steps {
		if c.status.Steps[i].ID == id {
			return &c.status.Steps[i]
		}
	}
	return nil
}
func (c *Client) dependenciesFinished(step Step) bool {
	for _, id := range step.DependsOn {
		if !stepSucceeded(c.stepStatus(id).Status) {
			return false
		}
	}
	if step.PlanFromStep != "" {
		plan := c.stepStatus(step.PlanFromStep)
		if plan == nil || plan.Status != string(models.AppRunnerJobStatusFinished) || plan.SourceRunID != "" {
			return false
		}
	}
	return true
}
func (c *Client) job(step Step) *models.AppRunnerJob {
	status := c.stepStatus(step.ID)
	return &models.AppRunnerJob{ID: step.ID, Type: models.AppRunnerJobType(step.JobType), Operation: models.AppRunnerJobOperationType(step.JobOperation), Group: models.AppRunnerJobGroup(step.JobGroup), Status: models.AppRunnerJobStatus(status.Status), LogStreamID: "airgap-" + step.ID, OrgID: c.envelope.OrgID, OwnerID: c.envelope.InstallID, ExecutionTimeout: int64(24 * time.Hour)}
}
func (c *Client) checkDoneLocked() {
	if c.result.FailedStep != "" {
		c.status.Status = statestore.RunStatusFailedPendingRetry
		c.status.FailedStep = c.result.FailedStep
		c.status.ResultDirective = statestore.DirectiveAwaitRetry
		if err := c.store.WriteStatus(c.status); err != nil {
			c.logger.Error("write pending-retry airgap status", zap.Error(err))
		}
		return
	}
	if c.status.ApprovalRequired {
		return
	}
	for _, step := range c.status.Steps {
		if !stepResolved(step.Status) {
			return
		}
	}
	c.result.Succeeded = true
	c.finalizeLocked(statestore.RunStatusFinished, "")
}

func stepSucceeded(status string) bool {
	return status == string(models.AppRunnerJobStatusFinished) || status == string(models.AppStatusAutoDashSkipped) || status == string(models.AppStatusUserDashSkipped)
}

func stepResolved(status string) bool {
	return stepSucceeded(status) || status == string(models.AppRunnerJobStatusNotDashAttempted)
}

func (c *Client) RunID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status.RunID
}

func (c *Client) ApplyControl(action string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if action != statestore.ControlActionCancel && c.status.Status != statestore.RunStatusFailedPendingRetry {
		return fmt.Errorf("run %s is not awaiting a retry decision", c.status.RunID)
	}
	if action == statestore.ControlActionCancel && c.status.Status != statestore.RunStatusInProgress && c.status.Status != statestore.RunStatusFailedPendingRetry {
		return fmt.Errorf("run %s is not cancellable", c.status.RunID)
	}
	switch action {
	case statestore.ControlActionRetry:
		now := time.Now().UTC()
		c.status.Status = statestore.RunStatusFailed
		c.status.ResultDirective = statestore.DirectiveRetryGroup
		c.status.FinishedAt = &now
		if err := c.store.WriteStatus(c.status); err != nil {
			return err
		}
		c.status = retryStatus(c.envelope, c.status, now)
		c.result = Result{}
		c.executions = map[string]*models.AppRunnerJobExecution{}
		c.applyApproved = !c.status.ApprovalRequired
		return c.store.WriteStatus(c.status)
	case statestore.ControlActionUserSkip:
		failed := c.stepStatus(c.status.FailedStep)
		if failed == nil {
			return fmt.Errorf("failed step not found")
		}
		failed.Status = string(models.AppStatusUserDashSkipped)
		failed.ResultDirective = statestore.DirectiveContinue
		if c.isPlanStep(failed.ID) {
			c.invalidatePlanDescendantsLocked(failed.ID)
			c.status.ApprovalRequired = false
			c.applyApproved = false
		}
		c.status.Status = statestore.RunStatusInProgress
		c.status.FailedStep = ""
		c.status.ResultDirective = statestore.DirectiveContinue
		c.result = Result{}
		if err := c.store.WriteStatus(c.status); err != nil {
			return err
		}
		c.checkDoneLocked()
		return nil
	case statestore.ControlActionCancel:
		now := time.Now().UTC()
		for i := range c.status.Steps {
			switch c.status.Steps[i].Status {
			case string(models.AppRunnerJobStatusInDashProgress):
				c.status.Steps[i].Status = string(models.AppRunnerJobStatusCancelled)
				c.status.Steps[i].FinishedAt = &now
			case string(models.AppRunnerJobStatusAvailable):
				c.status.Steps[i].Status = string(models.AppRunnerJobStatusNotDashAttempted)
			}
		}
		c.status.ResultDirective = statestore.DirectiveStop
		c.finalizeLocked(statestore.RunStatusCancelled, c.status.FailedStep)
		return nil
	default:
		return fmt.Errorf("unknown control action %q", action)
	}
}

func (c *Client) isPlanStep(id string) bool {
	for _, step := range c.envelope.Steps {
		if step.PlanFromStep == id {
			return true
		}
	}
	return false
}

func (c *Client) invalidatePlanDescendantsLocked(planID string) {
	invalid := map[string]bool{planID: true}
	for changed := true; changed; {
		changed = false
		for _, step := range c.envelope.Steps {
			blocked := step.PlanFromStep != "" && invalid[step.PlanFromStep]
			for _, dependency := range step.DependsOn {
				blocked = blocked || invalid[dependency]
			}
			if blocked && !invalid[step.ID] {
				invalid[step.ID], changed = true, true
			}
		}
	}
	delete(invalid, planID)
	for i := range c.status.Steps {
		if invalid[c.status.Steps[i].ID] {
			c.status.Steps[i].Status = string(models.AppRunnerJobStatusNotDashAttempted)
			c.status.Steps[i].ResultDirective = ""
		}
	}
}

// finalizeLocked runs after handlers persist all artifacts and only once per process.
func (c *Client) finalizeLocked(runStatus, failedStep string) {
	c.doneOnce.Do(func() {
		c.status.Status = runStatus
		c.status.FailedStep = failedStep
		if runStatus == statestore.RunStatusFinished && c.status.ResultDirective == "" {
			c.status.ResultDirective = statestore.DirectiveContinue
		}
		if failedStep != "" {
			for i := range c.status.Steps {
				if c.status.Steps[i].Status == string(models.AppRunnerJobStatusAvailable) {
					c.status.Steps[i].Status = string(models.AppRunnerJobStatusNotDashAttempted)
				}
			}
		}
		if c.status.FinishedAt == nil {
			now := time.Now().UTC()
			c.status.FinishedAt = &now
		}
		if err := c.store.WriteStatus(c.status); err != nil {
			c.logger.Error("write final airgap status", zap.Error(err))
		}
		report, err := BuildReport(c.envelope, c.status, c.store)
		if err != nil {
			c.logger.Error("build airgap report", zap.Error(err))
		} else if err := c.store.WriteReport(report); err != nil {
			c.logger.Error("write airgap report", zap.Error(err))
		}
		close(c.done)
	})
}
