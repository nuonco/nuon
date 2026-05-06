package sandboxhandler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	cockerrors "github.com/cockroachdb/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	logPeriod  = time.Second / 4
	totalSteps = 6
)

// Handler is a universal sandbox job handler that replaces the real handler
// when sandbox mode is active. It implements the jobs.JobHandler interface.
type Handler struct {
	sandboxCfg *Config
	apiClient  nuonrunner.Client
	cfg        *internal.Config
	shutdowner fx.Shutdowner

	job       *models.AppRunnerJob
	execution *models.AppRunnerJobExecution
}

func New(
	sandboxCfg *Config,
	apiClient nuonrunner.Client,
	cfg *internal.Config,
	shutdowner fx.Shutdowner,
	job *models.AppRunnerJob,
	execution *models.AppRunnerJobExecution,
) *Handler {
	return &Handler{
		sandboxCfg: sandboxCfg,
		apiClient:  apiClient,
		cfg:        cfg,
		shutdowner: shutdowner,
		job:        job,
		execution:  execution,
	}
}

func (h *Handler) Name() string {
	return "sandbox"
}

func (h *Handler) JobType() models.AppRunnerJobType {
	return h.job.Type
}

func (h *Handler) JobStatus() models.AppRunnerJobStatus {
	return h.job.Status
}

// Reset implements jobs.StatefulJobHandler.
func (h *Handler) Reset(ctx context.Context) error {
	return h.execStepForStep(ctx, "resetting")
}

func (h *Handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return h.execStepForStep(ctx, "fetching")
}

func (h *Handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return h.execStepForStep(ctx, "validate")
}

func (h *Handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return h.execStepForStep(ctx, "initialize")
}

func (h *Handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// Tag this handler's logger with the same semantic-convention attributes
	// the real per-tool deploy/sandbox handlers emit, so logs from sandbox-mode
	// orgs are filterable in the dashboard by nuon.tool / service.name even
	// though they're produced by the universal stub handler.
	if l, err := pkgctx.Logger(ctx); err == nil {
		tool, svc := toolAndServiceForJobType(job.Type)
		l = l.With(
			zap.String("service.name", svc),
			zap.String("nuon.tool", tool),
			zap.String("nuon.deploy.kind", "sandbox."+tool),
			zap.String("nuon.job.type", string(job.Type)),
			zap.String("nuon.job.operation", string(job.Operation)),
		)
		// SetLoggerWithSpan re-stamps ContextField(ctx) so the step span
		// (current span on ctx, opened in executeJob) is picked up by every
		// emit through the contextual logger inside execSandboxStep /
		// execStepForStep. Plain SetLogger would inherit the parent ctx field
		// — which already points at the step ctx — but we re-attach to make
		// the intent explicit and stay symmetric with op.Start.
		ctx = pkgctx.SetLoggerWithSpan(ctx, l)
	}

	if job.Type == models.AppRunnerJobTypeActionsDashWorkflow {
		return h.execActionSandboxStep(ctx, job)
	}
	return h.execSandboxStep(ctx, job)
}

// toolAndServiceForJobType maps a runner job.Type to the (nuon.tool,
// service.name) pair the dashboard log filters key off. Tool labels match the
// real deploy-mode handlers (terraform/helm/kubernetes_manifest/job/pulumi/
// sync_secrets/action) so a single filter value covers both real and
// sandbox-mode logs; service.name is namespaced under runner.sandbox.* so
// users can still distinguish them when needed.
func toolAndServiceForJobType(jobType models.AppRunnerJobType) (string, string) {
	switch jobType {
	case models.AppRunnerJobTypeTerraformDashDeploy,
		models.AppRunnerJobTypeRunnerDashTerraform,
		models.AppRunnerJobTypeSandboxDashTerraform,
		models.AppRunnerJobTypeSandboxDashTerraformDashPlan,
		models.AppRunnerJobTypeTerraformDashModuleDashBuild:
		return "terraform", "runner.sandbox.terraform"
	case models.AppRunnerJobTypeHelmDashChartDashDeploy,
		models.AppRunnerJobTypeRunnerDashHelm,
		models.AppRunnerJobTypeHelmDashChartDashBuild:
		return "helm", "runner.sandbox.helm"
	case models.AppRunnerJobTypeKubernetesDashManifestDashDeploy,
		models.AppRunnerJobTypeKubernetesDashManifestDashBuild:
		return "kubernetes_manifest", "runner.sandbox.kubernetes_manifest"
	case models.AppRunnerJobTypeJobDashDeploy:
		return "job", "runner.sandbox.job"
	case models.AppRunnerJobTypePulumiDashDeploy,
		models.AppRunnerJobTypeSandboxDashPulumi,
		models.AppRunnerJobTypePulumiDashBuild:
		return "pulumi", "runner.sandbox.pulumi"
	case models.AppRunnerJobTypeSandboxDashSyncDashSecrets:
		return "sync_secrets", "runner.sandbox.sync_secrets"
	case models.AppRunnerJobTypeActionsDashWorkflow:
		return "action", "runner.sandbox.action"
	case models.AppRunnerJobTypeDockerDashBuild,
		models.AppRunnerJobTypeContainerDashImageDashBuild:
		return "docker", "runner.sandbox.docker"
	case models.AppRunnerJobTypeOciDashSync:
		return "oci_sync", "runner.sandbox.oci_sync"
	default:
		return "sandbox", "runner.sandbox"
	}
}

func (h *Handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	return h.execStepForStep(ctx, "cleanup")
}

func (h *Handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	return nil
}

func (h *Handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	// Check for per-step failure at "outputs" step
	if err := h.execStepForStep(ctx, "outputs"); err != nil {
		return nil, err
	}

	outputs, err := h.sandboxOutputs(ctx)
	if err != nil {
		return nil, cockerrors.Wrap(err, "unable to get sandbox outputs")
	}

	// Write plan contents / execution results as side effects
	if err := h.writeSandboxResults(ctx); err != nil {
		return nil, err
	}

	return outputs, nil
}

// execStepForStep checks if the config has FailAtStep set and it matches the current step.
// If nothing special, just logs the step.
func (h *Handler) execStepForStep(ctx context.Context, stepName string) error {
	l, _ := pkgctx.Logger(ctx)

	cfg := h.sandboxCfg
	if cfg == nil {
		if l != nil {
			l.Info("sandbox: in handler step", zap.String("step", stepName))
		}
		return nil
	}

	jobType := string(h.job.Type)

	if cfg.FailAtStep != "" && cfg.FailAtStep == stepName {
		if l != nil {
			l.Error("sandbox: injecting failure at step",
				zap.String("step", stepName),
				zap.String("job_type", jobType),
			)
		}
		msg := cfg.ErrorMessage
		if msg == "" {
			msg = fmt.Sprintf("sandbox: failure injected at step %s", stepName)
		}
		return errors.New(msg)
	}

	if l != nil {
		l.Info("sandbox: in handler step", zap.String("step", stepName), zap.String("job_type", jobType))
	}
	return nil
}

// execSandboxStep runs the main sandbox execution simulation with duration and log lines.
func (h *Handler) execSandboxStep(ctx context.Context, job *models.AppRunnerJob) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	jobType := string(job.Type)
	cfg := h.sandboxCfg
	duration := h.cfg.SandboxJobDuration

	if cfg != nil {
		l.Info("sandbox-handler: config loaded",
			zap.String("job_type", jobType),
			zap.Duration("duration", cfg.Duration),
			zap.Bool("has_error", cfg.ErrorMessage != ""),
			zap.Int("log_lines", len(cfg.LogLines)),
		)
		if cfg.Duration > 0 {
			duration = cfg.Duration
		}

		if cfg.ErrorMessage != "" {
			l.Info("error message is set, exiting early")
			return errors.New(cfg.ErrorMessage)
		}

		if cfg.TriggerShutdown {
			l.Error("sandbox: trigger_shutdown enabled for job type", zap.String("job_type", jobType))
			h.shutdowner.Shutdown()
			return errors.New("sandbox: shutdown triggered for job type " + jobType)
		}

		if cfg.SleepDuration > 0 {
			l.Info("sandbox: sleeping before execution",
				zap.Duration("sleep_duration", cfg.SleepDuration),
				zap.String("job_type", jobType),
			)
			select {
			case <-time.After(cfg.SleepDuration):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if cfg.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
			defer cancel()
		}
	}

	l.Info("duration loaded from config",
		zap.Duration("duration", duration),
		zap.String("duration_human", duration.String()),
	)

	// Open one op.Tool span per simulated tool operation so sandbox-mode
	// jobs render in the dashboard span tree with the same shape as a real
	// deploy job. The simulatedOps table mirrors the op names wrapped by
	// real handlers in Phase 3 (terraform.plan, helm.upgrade, etc.); jobs
	// not in the table fall back to a single <tool>.exec span so we never
	// regress prior behavior.
	tool, _ := toolAndServiceForJobType(job.Type)
	ops := simulatedOps(job.Type, job.Operation)
	if len(ops) == 0 {
		ops = []simOp{{op: "exec", fraction: 1.0}}
	}

	hasCustomLogs := cfg != nil && len(cfg.LogLines) > 0
	customLogIdx := 0

	for _, sop := range ops {
		opDur := time.Duration(float64(duration) * sop.fraction)
		if opDur <= 0 {
			continue
		}
		if err := h.runSimOp(ctx, l, tool, sop, opDur, hasCustomLogs, cfg, &customLogIdx); err != nil {
			return err
		}
	}

	l.Info("sandbox job complete",
		zap.String("job_type", jobType),
		zap.Duration("duration", duration),
	)

	if cfg != nil && cfg.ErrorMessage != "" {
		l.Info("sandbox: error message set for job, returning error",
			zap.String("job_type", string(job.Type)),
			zap.String("error_message", cfg.ErrorMessage),
		)
		return errors.New(cfg.ErrorMessage)
	}

	return nil
}

// runSimOp opens an op.Tool span for one simulated tool operation, sleeps
// for `dur` while emitting log lines on a ticker, and closes the span. If
// the parent ctx is cancelled the span is closed with an error so the
// dashboard tree colors it red.
//
// Custom log lines from sandbox config (cfg.LogLines) are consumed across op
// spans via *customLogIdx so the existing "use custom logs until exhausted,
// then fall back to canned output" semantics are preserved across the new
// multi-span layout.
func (h *Handler) runSimOp(
	ctx context.Context,
	parentLog *zap.Logger,
	tool string,
	sop simOp,
	dur time.Duration,
	useCustomLogs bool,
	cfg *Config,
	customLogIdx *int,
) error {
	opCtx, end := op.Tool(ctx, tool, sop.op)
	opLog := pkgctx.LoggerOrDefault(opCtx, parentLog)

	timeout := time.NewTimer(dur)
	defer timeout.Stop()
	ticker := time.NewTicker(logPeriod)
	defer ticker.Stop()

	cannedIdx := 0
	for {
		select {
		case <-opCtx.Done():
			err := fmt.Errorf("sandbox: op %s.%s cancelled", tool, sop.op)
			end(err)
			return err
		case <-ticker.C:
			switch {
			case useCustomLogs && cfg != nil && *customLogIdx < len(cfg.LogLines):
				opLog.Info(cfg.LogLines[*customLogIdx])
				*customLogIdx++
			case cannedIdx < len(sop.logs):
				opLog.Info(sop.logs[cannedIdx])
				cannedIdx++
			default:
				opLog.Info("(simulated) " + tool + "." + sop.op + " progressing")
			}
		case <-timeout.C:
			end(nil)
			return nil
		}
	}
}

// execActionSandboxStep handles sandbox mode for actions-workflow job types.
func (h *Handler) execActionSandboxStep(ctx context.Context, job *models.AppRunnerJob) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Check sandbox config failure modes before executing
	cfg := h.sandboxCfg
	if cfg != nil {
		if cfg.ErrorMessage != "" {
			l.Info("sandbox: error message set for action workflow, returning error",
				zap.String("job_type", string(job.Type)),
				zap.String("error_message", cfg.ErrorMessage),
			)
			return errors.New(cfg.ErrorMessage)
		}

		if cfg.TriggerShutdown {
			l.Error("sandbox: trigger_shutdown enabled for action workflow",
				zap.String("job_type", string(job.Type)),
			)
			h.shutdowner.Shutdown()
			return errors.New("sandbox: shutdown triggered for job type " + string(job.Type))
		}
	}

	l.Info("fetching actions job plan")
	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, job.ID)
	if err != nil {
		return cockerrors.Wrap(err, "unable to get job plan")
	}

	var plan plantypes.ActionWorkflowRunPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return cockerrors.Wrap(err, "unable to parse action workflow run plan")
	}

	run, err := h.apiClient.GetInstallActionWorkflowRun(ctx, plan.InstallID, plan.ID)
	if err != nil {
		return cockerrors.Wrap(err, "unable to get action workflow run")
	}

	isAdhoc := run.ActionWorkflowConfigID == ""

	var actionCfg *models.AppActionWorkflowConfig
	if !isAdhoc {
		l.Info("fetching actions workflow config")
		actionCfg, err = h.apiClient.GetActionWorkflowConfig(ctx, run.ActionWorkflowConfigID)
		if err != nil {
			return cockerrors.Wrap(err, "unable to get action workflow config")
		}
	}

	for idx, step := range run.Steps {
		var stepName string
		var actionWorkflowID string

		if isAdhoc {
			if step.AdhocConfig != nil {
				stepName = step.AdhocConfig.Name
			} else {
				stepName = "adhoc step"
			}
			actionWorkflowID = run.ID
		} else {
			stepCfg := actionCfg.Steps[idx]
			stepName = stepCfg.Name
			actionWorkflowID = actionCfg.ActionWorkflowID
		}

		l = l.With(
			zap.String("workflow_step_name", stepName),
			zap.String("step_run_id", step.ID),
		)

		l.Info(fmt.Sprintf("executing step %s (%d of %d)", stepName, idx+1, len(run.Steps)))

		_, err := h.apiClient.UpdateInstallActionWorkflowRunStep(ctx, plan.InstallID, actionWorkflowID, step.ID, &models.ServiceUpdateInstallActionWorkflowRunStepRequest{
			Status:            models.AppInstallActionWorkflowRunStepStatusFinished,
			ExecutionDuration: int64(time.Second * 5),
		})
		if err != nil {
			return cockerrors.Wrap(err, "unable to update step status")
		}
	}

	return nil
}

// sandboxOutputs returns the outputs map for a sandbox job.
func (h *Handler) sandboxOutputs(ctx context.Context) (map[string]interface{}, error) {
	if h.sandboxCfg != nil && len(h.sandboxCfg.Outputs) > 0 {
		return h.sandboxCfg.Outputs, nil
	}

	plan, err := h.getSandboxModePlan(ctx)
	if err != nil {
		return nil, cockerrors.Wrap(err, "unable to get sandbox mode plan")
	}

	if plan.SandboxMode == nil || !plan.SandboxMode.Enabled {
		return map[string]interface{}{}, nil
	}

	return plan.SandboxMode.Outputs, nil
}

// writeSandboxResults writes plan contents and execution results for sandbox jobs.
func (h *Handler) writeSandboxResults(ctx context.Context) error {
	if h.sandboxCfg != nil && h.sandboxCfg.PlanContents != "" {
		req := &models.ServiceCreateRunnerJobExecutionResultRequest{
			ContentsCompressed: compress(h.sandboxCfg.PlanContents),
			Success:            true,
		}
		if h.sandboxCfg.PlanDisplayContents != "" {
			req.ContentsDisplayCompressed = compress(h.sandboxCfg.PlanDisplayContents)
		}
		if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.job.ID, h.execution.ID, req); err != nil {
			return cockerrors.Wrap(err, "unable to write sandbox config plan contents")
		}
		return nil
	}

	// Fall back to plan-based sandbox mode outputs
	plan, err := h.getSandboxModePlan(ctx)
	if err != nil {
		return cockerrors.Wrap(err, "unable to get sandbox mode plan")
	}

	if plan.SandboxMode != nil && plan.SandboxMode.Terraform != nil {
		if err := h.writeTerraformSandboxMode(ctx, plan.SandboxMode.Terraform); err != nil {
			return cockerrors.Wrap(err, "unable to write sandbox mode terraform")
		}
	}
	if plan.SandboxMode != nil && plan.SandboxMode.Helm != nil {
		if err := h.writeHelmSandboxMode(ctx, plan.SandboxMode.Helm); err != nil {
			return cockerrors.Wrap(err, "unable to write sandbox mode helm")
		}
	}
	if plan.SandboxMode != nil && plan.SandboxMode.KubernetesManifest != nil {
		if err := h.writeKubernetesManifestSandboxMode(ctx, plan.SandboxMode.KubernetesManifest); err != nil {
			return cockerrors.Wrap(err, "unable to write sandbox mode kubernetes_manifest")
		}
	}
	if plan.SandboxMode != nil && plan.SandboxMode.Pulumi != nil {
		if err := h.writePulumiSandboxMode(ctx, plan.SandboxMode.Pulumi); err != nil {
			return cockerrors.Wrap(err, "unable to write sandbox mode pulumi")
		}
	}

	return nil
}

func (h *Handler) getSandboxModePlan(ctx context.Context) (*plantypes.MinSandboxMode, error) {
	var plan plantypes.MinSandboxMode

	planJSON, err := h.apiClient.GetJobPlanJSON(ctx, h.job.ID)
	if err != nil {
		return nil, cockerrors.Wrap(err, "unable to get job plan")
	}

	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil, cockerrors.Wrap(err, "unable to convert to sandbox plan")
	}

	return &plan, nil
}

func (h *Handler) writeTerraformSandboxMode(ctx context.Context, plan *plantypes.TerraformSandboxMode) error {
	params := url.Values{
		"job_id":       {h.job.ID},
		"workspace_id": {plan.WorkspaceID},
		"token":        {h.cfg.RunnerAPIToken},
	}

	u, err := url.JoinPath(h.cfg.RunnerAPIURL, "/v1/terraform-backend")
	if err != nil {
		return cockerrors.Wrap(err, "unable to get url")
	}
	u = u + "?" + params.Encode()

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(plan.StateJSON))
	if err != nil {
		return cockerrors.Wrap(err, "unable to create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cockerrors.Wrap(err, "unable to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cockerrors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if _, err := h.apiClient.UpdateTerraformStateJSON(ctx, plan.WorkspaceID, &h.job.ID, []byte(plan.StateJSON)); err != nil {
		return cockerrors.Errorf("unable to update state json")
	}

	if len(plan.PlanContents) > 0 {
		var planDisplayJson *map[string]interface{}
		err = json.Unmarshal([]byte(plan.PlanDisplayContents), &planDisplayJson)
		if err != nil {
			return cockerrors.Wrap(err, "unable to unmarshal plan display")
		}

		if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.job.ID, h.execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			ContentsCompressed:        compress(plan.PlanContents),
			ContentsDisplayCompressed: compress(plan.PlanDisplayContents),
			Success:                   true,
		}); err != nil {
			return cockerrors.Wrap(err, "unable to create job execution results")
		}
	}

	return nil
}

func (h *Handler) writeHelmSandboxMode(ctx context.Context, plan *plantypes.HelmSandboxMode) error {
	if len(plan.PlanContents) > 0 {
		var planDisplayJson *map[string]interface{}
		err := json.Unmarshal([]byte(plan.PlanDisplayContents), &planDisplayJson)
		if err != nil {
			return cockerrors.Wrap(err, "unable to unmarshal plan display")
		}

		h.apiClient.CreateJobExecutionResult(ctx, h.job.ID, h.execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			ContentsCompressed:        compress(plan.PlanContents),
			ContentsDisplayCompressed: compress(plan.PlanDisplayContents),
		})
	}

	return nil
}

func (h *Handler) writeKubernetesManifestSandboxMode(ctx context.Context, plan *plantypes.KubernetesSandboxMode) error {
	if len(plan.PlanContents) > 0 {
		var planDisplayJson *map[string]interface{}
		err := json.Unmarshal([]byte(plan.PlanDisplayContents), &planDisplayJson)
		if err != nil {
			return cockerrors.Wrap(err, "unable to unmarshal plan display")
		}

		h.apiClient.CreateJobExecutionResult(ctx, h.job.ID, h.execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			ContentsCompressed:        compress(plan.PlanContents),
			ContentsDisplayCompressed: compress(plan.PlanDisplayContents),
		})
	}

	return nil
}

func (h *Handler) writePulumiSandboxMode(ctx context.Context, plan *plantypes.PulumiSandboxMode) error {
	if len(plan.PlanContents) > 0 {
		if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.job.ID, h.execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			ContentsCompressed:        compress(plan.PlanContents),
			ContentsDisplayCompressed: compress(plan.PlanDisplayContents),
		}); err != nil {
			return cockerrors.Wrap(err, "unable to create job execution result")
		}
	}

	return nil
}

func compress(s string) string {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte(s))
	gz.Close()
	b64 := base64.URLEncoding.EncodeToString(b.Bytes())
	return b64
}
