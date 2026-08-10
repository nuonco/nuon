package controlplanejob

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/pkg/metrics"
	runnercontrolplane "github.com/nuonco/nuon/pkg/runner/controlplane"
	"github.com/nuonco/nuon/pkg/runner/errcapture"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/aws"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/generic"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/helm"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/terraform"
	runnerhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Activities struct {
	db               *gorm.DB
	chDB             *gorm.DB
	cfg              *internal.Config
	l                *zap.Logger
	mw               metrics.Writer
	blobSvc          blobstore.Service
	statusActivities *statusactivities.Activities
	kafka            *kafka.Producer
}

type ActivityParams struct {
	fx.In

	DB   *gorm.DB `name:"psql"`
	CHDB *gorm.DB `name:"ch"`
	Cfg  *internal.Config
	L    *zap.Logger
	MW   metrics.Writer

	BlobSvc          blobstore.Service
	StatusActivities *statusactivities.Activities
	Kafka            *kafka.Producer
}

func NewActivities(params ActivityParams) *Activities {
	return &Activities{
		db:               params.DB,
		chDB:             params.CHDB,
		cfg:              params.Cfg,
		l:                params.L,
		mw:               params.MW,
		blobSvc:          params.BlobSvc,
		statusActivities: params.StatusActivities,
		kafka:            params.Kafka,
	}
}

func (a *Activities) AllActivities() []any {
	return []any{
		a.ControlPlaneJobEnsureExecution,
		a.ControlPlaneJobRunJob,
		a.ControlPlaneJobFinalize,
	}
}

type EnsureExecutionResponse struct {
	ExecutionID         string        `json:"execution_id"`
	JobExecutionTimeout time.Duration `json:"job_execution_timeout"`
}

// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix ControlPlaneJob
func (a *Activities) ensureExecution(ctx context.Context, jobID string) (*EnsureExecutionResponse, error) {
	var job app.RunnerJob
	var execution app.RunnerJobExecution
	if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Scopes(scopes.WithDisableViews).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(&app.RunnerJob{ID: jobID}).
			First(&job).Error; err != nil {
			return fmt.Errorf("unable to get runner job: %w", err)
		}
		if job.Executor != app.RunnerJobExecutorControlPlane {
			return fmt.Errorf("runner job %s has executor %q, not %q", job.ID, job.Executor, app.RunnerJobExecutorControlPlane)
		}
		if job.Status.IsTerminal() {
			return fmt.Errorf("runner job %s is already %s", job.ID, job.Status)
		}

		res := tx.
			Where(&app.RunnerJobExecution{RunnerJobID: job.ID}).
			Order("created_at desc").
			First(&execution)
		if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to get existing job execution: %w", res.Error)
		}
		if res.Error == gorm.ErrRecordNotFound || !execution.Status.IsRunning() {
			execution = app.RunnerJobExecution{RunnerJobID: job.ID, Status: app.RunnerJobExecutionStatusPending}
			if err := tx.Create(&execution).Error; err != nil {
				return fmt.Errorf("unable to create runner job execution: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)
	if job.StartedAt.IsZero() {
		if err := a.db.WithContext(ctx).
			Model(&app.RunnerJob{ID: job.ID}).
			Update("started_at", time.Now().UTC()).Error; err != nil {
			return nil, fmt.Errorf("unable to update runner job started_at: %w", err)
		}
	}
	if err := a.updateJob(ctx, job.ID, app.RunnerJobStatusInProgress, "in-progress"); err != nil {
		return nil, err
	}
	return &EnsureExecutionResponse{ExecutionID: execution.ID, JobExecutionTimeout: job.ExecutionTimeout}, nil
}

// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix ControlPlaneJob
func (a *Activities) runJob(ctx context.Context, jobID, executionID string) error {
	activity.RecordHeartbeat(ctx, "starting")
	stopHeartbeat := runnercontrolplane.HeartbeatUntilDone(ctx, func() { activity.RecordHeartbeat(ctx, "running") })
	defer stopHeartbeat()

	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	var execution app.RunnerJobExecution
	if err := a.db.WithContext(ctx).Where(&app.RunnerJobExecution{ID: executionID, RunnerJobID: jobID}).First(&execution).Error; err != nil {
		return fmt.Errorf("unable to get runner job execution: %w", err)
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)
	capture := errcapture.New()
	ctx = errcapture.NewContext(ctx, capture)

	executor, err := runnercontrolplane.NewExecutor(a, a.l, runnercontrolplane.Config{GitRef: a.cfg.GitRef})
	if err != nil {
		return fmt.Errorf("unable to create control-plane executor: %w", err)
	}
	if err := executor.Execute(ctx, toRunnerJobModel(job), toRunnerExecutionModel(&execution)); err != nil {
		if failureErr := a.ensureFailureResult(ctx, jobID, executionID, err.Error()); failureErr != nil {
			a.l.Warn("unable to persist control-plane failure result", zap.Error(failureErr))
		}
		if errors.Is(err, context.Canceled) {
			return temporal.NewCanceledError()
		}
		return err
	}
	return nil
}

type FinalizeOutcome struct {
	Status app.RunnerJobExecutionStatus `json:"status"`
	Error  string                       `json:"error,omitempty"`
}

type FinalizeResponse struct {
	Status app.RunnerJobExecutionStatus `json:"status"`
}

// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix ControlPlaneJob
func (a *Activities) finalize(ctx context.Context, jobID, executionID string, outcome FinalizeOutcome) (*FinalizeResponse, error) {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)

	var execution app.RunnerJobExecution
	if err := a.db.WithContext(ctx).
		Where(&app.RunnerJobExecution{ID: executionID, RunnerJobID: jobID}).
		First(&execution).Error; err != nil {
		return nil, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	execStatus := outcome.Status
	if execStatus == "" {
		execStatus = app.RunnerJobExecutionStatusFinished
	}
	if !execution.Status.IsRunning() {
		execStatus = execution.Status
	}

	description := "finished"
	if execStatus != app.RunnerJobExecutionStatusFinished {
		description = outcome.Error
		if description == "" {
			description = string(execStatus)
		}
		if err := a.ensureFailureResult(ctx, jobID, executionID, outcome.Error); err != nil {
			return nil, err
		}
	}
	if _, err := a.UpdateJobExecution(ctx, jobID, executionID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatus(execStatus), StatusDescription: description}); err != nil {
		return nil, err
	}
	if err := a.db.WithContext(ctx).Where(&app.RunnerJobExecution{ID: executionID, RunnerJobID: jobID}).First(&execution).Error; err != nil {
		return nil, fmt.Errorf("unable to refresh runner job execution: %w", err)
	}
	jobStatus := jobStatusForExecution(execution.Status)
	if err := a.updateJob(ctx, jobID, jobStatus, description); err != nil {
		return nil, err
	}
	job, err = a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	effectiveStatus := execution.Status
	if job.Status.IsTerminal() && job.Status != app.RunnerJobStatusFinished {
		effectiveStatus = executionStatusForJob(job.Status)
	}
	if err := a.db.WithContext(ctx).
		Model(&app.RunnerJob{ID: jobID}).
		Update("finished_at", time.Now().UTC()).Error; err != nil {
		return nil, fmt.Errorf("unable to update runner job finished_at: %w", err)
	}
	return &FinalizeResponse{Status: effectiveStatus}, nil
}

func jobStatusForExecution(status app.RunnerJobExecutionStatus) app.RunnerJobStatus {
	switch status {
	case app.RunnerJobExecutionStatusFinished:
		return app.RunnerJobStatusFinished
	case app.RunnerJobExecutionStatusTimedOut:
		return app.RunnerJobStatusTimedOut
	case app.RunnerJobExecutionStatusCancelled:
		return app.RunnerJobStatusCancelled
	case app.RunnerJobExecutionStatusNotAttempted:
		return app.RunnerJobStatusNotAttempted
	default:
		return app.RunnerJobStatusFailed
	}
}

func (a *Activities) GetJobPlanJSON(ctx context.Context, jobID string) (string, error) {
	var plan app.RunnerJobPlan
	if err := a.db.WithContext(ctx).Where(&app.RunnerJobPlan{RunnerJobID: jobID}).First(&plan).Error; err != nil {
		return "", fmt.Errorf("unable to get runner job plan: %w", err)
	}
	return plan.PlanJSON, nil
}

func (a *Activities) GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error) {
	var plan app.RunnerJobPlan
	if err := a.db.WithContext(ctx).Where(&app.RunnerJobPlan{RunnerJobID: jobID}).First(&plan).Error; err != nil {
		return nil, fmt.Errorf("unable to get runner job plan: %w", err)
	}

	compositePlan, _ := plan.GetCompositePlan(ctx, a.cfg.BlobReadEnabled)
	if compositePlan.IsEmpty() {
		job, err := a.getJob(ctx, jobID)
		if err != nil {
			return nil, err
		}
		compositePlan, err = plan.DeriveCompositePlan(job)
		if err != nil {
			return nil, err
		}
	}

	contents, err := json.Marshal(compositePlan)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal composite plan: %w", err)
	}
	var result models.PlantypesCompositePlan
	if err := json.Unmarshal(contents, &result); err != nil {
		return nil, fmt.Errorf("unable to convert composite plan: %w", err)
	}
	return &result, nil
}

func (a *Activities) UpdateJobExecution(ctx context.Context, jobID, executionID string, req *models.ServiceUpdateRunnerJobExecutionRequest) (*models.AppRunnerJobExecution, error) {
	var execution app.RunnerJobExecution
	requestedStatus := app.RunnerJobExecutionStatus(req.Status)
	if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job app.RunnerJob
		if err := tx.
			Scopes(scopes.WithDisableViews).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(&app.RunnerJob{ID: jobID}).
			First(&job).Error; err != nil {
			return fmt.Errorf("unable to get runner job: %w", err)
		}
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(&app.RunnerJobExecution{ID: executionID, RunnerJobID: jobID}).
			First(&execution).Error; err != nil {
			return fmt.Errorf("unable to get runner job execution: %w", err)
		}
		if !execution.Status.IsRunning() {
			return nil
		}

		effectiveStatus := requestedStatus
		if job.Status.IsTerminal() {
			effectiveStatus = executionStatusForJob(job.Status)
		}
		if err := tx.Model(&execution).Update("status", effectiveStatus).Error; err != nil {
			return fmt.Errorf("unable to update runner job execution: %w", err)
		}
		execution.Status = effectiveStatus
		return nil
	}); err != nil {
		return nil, err
	}
	runnerhelpers.AuditJobExecutionResult(ctx, a.db, a.mw, execution.ID, execution.Status, "control_plane")

	description := req.StatusDescription
	if execution.Status != requestedStatus {
		description = string(execution.Status)
	}
	if err := a.reconcileExecutionStatusV2(ctx, &execution, description); err != nil {
		return nil, err
	}
	return toRunnerExecutionModel(&execution), nil
}

func executionStatusForJob(status app.RunnerJobStatus) app.RunnerJobExecutionStatus {
	switch status {
	case app.RunnerJobStatusFinished:
		return app.RunnerJobExecutionStatusFinished
	case app.RunnerJobStatusTimedOut:
		return app.RunnerJobExecutionStatusTimedOut
	case app.RunnerJobStatusCancelled:
		return app.RunnerJobExecutionStatusCancelled
	case app.RunnerJobStatusNotAttempted:
		return app.RunnerJobExecutionStatusNotAttempted
	default:
		return app.RunnerJobExecutionStatusFailed
	}
}

func (a *Activities) reconcileExecutionStatusV2(ctx context.Context, execution *app.RunnerJobExecution, description string) error {
	if execution.StatusV2.Status == app.Status(execution.Status) {
		return nil
	}
	if description == "" {
		description = string(execution.Status)
	}
	if err := a.statusActivities.UpdateRunnerJobExecutionStatusV2(ctx, statusactivities.UpdateRunnerJobExecutionStatusV2Request{
		RunnerJobExecutionID: execution.ID,
		Status:               execution.Status,
		StatusDescription:    truncateStatusDescription(description),
	}); err != nil {
		return fmt.Errorf("unable to update runner job execution status history: %w", err)
	}
	return nil
}

func (a *Activities) CreateJobExecutionResult(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)

	resolveProvider := func() errparse.Provider {
		return errparse.ResolveRunnerJobProvider(ctx, a.db, job)
	}
	compositeError, err := errparse.ParseRunnerJobResult(req.Success, req.ErrorMetadata, job, resolveProvider)
	if err != nil {
		a.l.Warn("unable to build composite error; omitting enrichment",
			zap.String("runner_job_id", job.ID),
			zap.Error(err))
	}
	result := app.RunnerJobExecutionResult{
		OrgID:                job.OrgID,
		RunnerJobExecutionID: executionID,
		Success:              req.Success,
		Contents:             req.Contents,
		ErrorCode:            int(req.ErrorCode),
		ErrorMetadata:        stringMapToHstore(req.ErrorMetadata),
		CompositeError:       compositeError,
	}
	if req.ContentsCompressed != "" {
		contents, err := base64.URLEncoding.DecodeString(req.ContentsCompressed)
		if err != nil {
			return nil, fmt.Errorf("unable to decode compressed contents: %w", err)
		}
		result.ContentsGzip = contents
	}
	if req.ContentsDisplayCompressed != "" {
		contents, err := base64.URLEncoding.DecodeString(req.ContentsDisplayCompressed)
		if err != nil {
			return nil, fmt.Errorf("unable to decode compressed contents display: %w", err)
		}
		result.ContentsDisplayGzip = contents
	}
	if req.ContentsDisplay != nil {
		byts, err := json.Marshal(req.ContentsDisplay)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal contents display: %w", err)
		}
		result.ContentsDisplay = byts
	}

	persisted, created, err := runnerhelpers.CreateJobExecutionResultIfAbsent(ctx, a.db, &result)
	if err != nil {
		return nil, err
	}
	if created && !req.Success {
		matched := "miss"
		if result.CompositeError != nil && result.CompositeError.Type != "" {
			matched = string(result.CompositeError.Type)
		}
		tool := string(errparse.ToolForRunnerJob(job))
		if tool == "" {
			tool = "unknown"
		}
		group := string(job.Group)
		if group == "" {
			group = "unknown"
		}
		a.mw.Incr("runner.composite_error_parse", []string{"tool:" + tool, "group:" + group, "matched_type:" + matched})
	}

	if created {
		if err := a.applyComponentBuildSourceIdentity(ctx, job, req); err != nil {
			a.l.Warn("unable to apply component build source identity", zap.Error(err))
		}
		a.refreshOwnerCompositeError(ctx, job, result.CompositeError)
	}

	return &models.AppRunnerJobExecutionResult{ID: persisted.ID, Success: persisted.Success}, nil
}

func (a *Activities) refreshOwnerCompositeError(ctx context.Context, job *app.RunnerJob, compositeError any) {
	if job.OwnerID == "" {
		return
	}

	var res *gorm.DB
	switch job.OwnerType {
	case "install_deploys":
		res = a.db.WithContext(ctx).Model(&app.InstallDeploy{ID: job.OwnerID}).Update("composite_error", compositeError)
	case "install_sandbox_runs":
		res = a.db.WithContext(ctx).Model(&app.InstallSandboxRun{ID: job.OwnerID}).Update("composite_error", compositeError)
	default:
		return
	}
	if res.Error != nil {
		a.l.Warn("unable to refresh owner composite error", zap.String("owner_type", job.OwnerType), zap.String("owner_id", job.OwnerID), zap.Error(res.Error))
	}
}

func (a *Activities) applyComponentBuildSourceIdentity(ctx context.Context, job *app.RunnerJob, req *models.ServiceCreateRunnerJobExecutionResultRequest) error {
	if req.SourceDigest == "" || job.OwnerType != "component_builds" || job.OwnerID == "" {
		return nil
	}

	var resolvedAt *time.Time
	if req.ResolvedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, req.ResolvedAt)
		if err != nil {
			return fmt.Errorf("unable to parse resolved_at: %w", err)
		}
		resolvedAt = &parsed
	}

	build := app.ComponentBuild{ID: job.OwnerID}
	updates := app.ComponentBuild{
		SourceRef:       req.SourceRef,
		SourceImage:     req.SourceImage,
		ResolvedTag:     req.ResolvedTag,
		SourceDigest:    req.SourceDigest,
		SourceMediaType: req.SourceMediaType,
		ResolvedAt:      resolvedAt,
		NoOp:            req.NoOp,
	}
	if err := a.db.WithContext(ctx).
		Model(&build).
		Select("source_ref", "source_image", "resolved_tag", "source_digest", "source_media_type", "resolved_at", "no_op").
		Updates(updates).Error; err != nil {
		return fmt.Errorf("unable to update component build source identity: %w", err)
	}
	return nil
}

func (a *Activities) CreateJobExecutionOutputs(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error) {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)
	ctx = blobstore.WithBlobService(ctx, a.blobSvc)
	ctx = blobstore.WithBlobWriteEnabled(ctx, true)
	byts, err := json.Marshal(req.Outputs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal outputs: %w", err)
	}
	outputs := app.RunnerJobExecutionOutputs{OrgID: job.OrgID, RunnerJobExecutionID: executionID, Outputs: byts, OutputsBlob: &blobstore.Blob{}}
	outputs.OutputsBlob.Set(string(byts))

	var existing app.RunnerJobExecutionOutputs
	res := a.db.WithContext(ctx).Where(&app.RunnerJobExecutionOutputs{RunnerJobExecutionID: executionID}).First(&existing)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get existing runner job execution outputs: %w", res.Error)
	}
	if res.Error == gorm.ErrRecordNotFound {
		if err := a.db.WithContext(ctx).Create(&outputs).Error; err != nil {
			return nil, fmt.Errorf("unable to create runner job execution outputs: %w", err)
		}
		return &models.AppRunnerJobExecutionOutputs{ID: outputs.ID}, nil
	}

	outputsBlob := &blobstore.Blob{}
	outputsBlob.Set(string(byts))
	if err := outputsBlob.BeforeCreate(a.db.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("unable to write runner job execution outputs blob: %w", err)
	}
	if err := a.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"org_id":       outputs.OrgID,
		"outputs":      outputs.Outputs,
		"outputs_blob": outputsBlob,
	}).Error; err != nil {
		return nil, fmt.Errorf("unable to update runner job execution outputs: %w", err)
	}
	return &models.AppRunnerJobExecutionOutputs{ID: existing.ID}, nil
}

func (a *Activities) ensureFailureResult(ctx context.Context, jobID, executionID, message string) error {
	var existing app.RunnerJobExecutionResult
	res := a.db.WithContext(ctx).Where(&app.RunnerJobExecutionResult{RunnerJobExecutionID: executionID}).First(&existing)
	if res.Error == nil {
		return nil
	}
	if res.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("unable to get existing runner job execution result: %w", res.Error)
	}

	_, err := a.CreateJobExecutionResult(ctx, jobID, executionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:   false,
		ErrorCode: 0,
		ErrorMetadata: map[string]string{
			"handler":      "control-plane",
			"message":      message,
			"error_output": errcapture.Output(ctx),
		},
	})
	return err
}

func (a *Activities) getJob(ctx context.Context, jobID string) (*app.RunnerJob, error) {
	var job app.RunnerJob
	if err := a.db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(&app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", err)
	}
	return &job, nil
}

func (a *Activities) GetJob(ctx context.Context, jobID string) (*models.AppRunnerJob, error) {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return toRunnerJobModel(job), nil
}

func (a *Activities) updateJob(ctx context.Context, jobID string, status app.RunnerJobStatus, description string) error {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Status.IsTerminal() {
		if job.Status != status {
			description = string(job.Status)
		}
		return a.reconcileJobStatusV2(ctx, job, description)
	}
	res := a.db.WithContext(ctx).
		Model(&app.RunnerJob{}).
		Where(&app.RunnerJob{ID: jobID, Status: job.Status}).
		Updates(app.RunnerJob{Status: status, StatusDescription: description})
	if res.Error != nil {
		return fmt.Errorf("unable to update runner job status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		job, err = a.getJob(ctx, jobID)
		if err != nil {
			return err
		}
		if job.Status != status {
			description = string(job.Status)
		}
		return a.reconcileJobStatusV2(ctx, job, description)
	}
	if err := a.statusActivities.UpdateRunnerJobStatusV2(ctx, statusactivities.UpdateRunnerJobStatusV2Request{
		RunnerJobID:       jobID,
		Status:            status,
		StatusDescription: truncateStatusDescription(description),
	}); err != nil {
		return fmt.Errorf("unable to update runner job status history: %w", err)
	}
	return nil
}

func (a *Activities) reconcileJobStatusV2(ctx context.Context, job *app.RunnerJob, description string) error {
	if job.StatusV2.Status == app.Status(job.Status) {
		return nil
	}
	if description == "" {
		description = string(job.Status)
	}
	if err := a.statusActivities.UpdateRunnerJobStatusV2(ctx, statusactivities.UpdateRunnerJobStatusV2Request{
		RunnerJobID:       job.ID,
		Status:            job.Status,
		StatusDescription: truncateStatusDescription(description),
	}); err != nil {
		return fmt.Errorf("unable to update runner job status history: %w", err)
	}
	return nil
}

func toRunnerJobModel(job *app.RunnerJob) *models.AppRunnerJob {
	logStreamID := ""
	if job.LogStreamID != nil {
		logStreamID = *job.LogStreamID
	}
	return &models.AppRunnerJob{ID: job.ID, RunnerID: job.RunnerID, OwnerID: job.OwnerID, OwnerType: job.OwnerType, LogStreamID: logStreamID, ExecutionTimeout: int64(job.ExecutionTimeout), Status: models.AppRunnerJobStatus(job.Status), Type: models.AppRunnerJobType(job.Type), Group: models.AppRunnerJobGroup(job.Group), Operation: models.AppRunnerJobOperationType(job.Operation), OrgID: job.OrgID, Metadata: hstoreToMap(job.Metadata)}
}

func toRunnerExecutionModel(execution *app.RunnerJobExecution) *models.AppRunnerJobExecution {
	return &models.AppRunnerJobExecution{ID: execution.ID, RunnerJobID: execution.RunnerJobID, Status: models.AppRunnerJobExecutionStatus(execution.Status), OrgID: execution.OrgID, Metadata: hstoreToMap(execution.Metadata)}
}

func hstoreToMap(h pgtype.Hstore) map[string]string {
	out := map[string]string{}
	for key, val := range h {
		if val != nil {
			out[key] = *val
		}
	}
	return out
}

func stringMapToHstore(in map[string]string) pgtype.Hstore {
	out := pgtype.Hstore{}
	for key, val := range in {
		v := val
		out[key] = &v
	}
	return out
}

const statusDescriptionMaxLen = 2048

func truncateStatusDescription(s string) string {
	if len(s) <= statusDescriptionMaxLen {
		return s
	}
	return s[:statusDescriptionMaxLen] + "…(truncated)"
}
