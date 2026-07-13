package controlplanejob

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	runnercontrolplane "github.com/nuonco/nuon/pkg/runner/controlplane"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type Activities struct {
	db   *gorm.DB
	chDB *gorm.DB
	cfg  *internal.Config
	l    *zap.Logger
}

type ActivityParams struct {
	fx.In

	DB   *gorm.DB `name:"psql"`
	CHDB *gorm.DB `name:"ch"`
	Cfg  *internal.Config
	L    *zap.Logger
}

func NewActivities(params ActivityParams) *Activities {
	return &Activities{db: params.DB, chDB: params.CHDB, cfg: params.Cfg, l: params.L}
}

type EnsureExecutionRequest struct {
	JobID string `json:"job_id" validate:"required"`
}

type EnsureExecutionResponse struct {
	ExecutionID         string        `json:"execution_id"`
	JobExecutionTimeout time.Duration `json:"job_execution_timeout"`
}

func (a *Activities) EnsureExecution(ctx context.Context, req *EnsureExecutionRequest) (*EnsureExecutionResponse, error) {
	job, err := a.getJob(ctx, req.JobID)
	if err != nil {
		return nil, err
	}
	if job.Executor != app.RunnerJobExecutorControlPlane {
		return nil, fmt.Errorf("runner job %s has executor %q, not %q", job.ID, job.Executor, app.RunnerJobExecutorControlPlane)
	}

	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)

	var execution app.RunnerJobExecution
	res := a.db.WithContext(ctx).
		Where(&app.RunnerJobExecution{RunnerJobID: job.ID}).
		Order("created_at desc").
		First(&execution)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get existing job execution: %w", res.Error)
	}
	if res.Error == gorm.ErrRecordNotFound || !execution.Status.IsRunning() {
		execution = app.RunnerJobExecution{RunnerJobID: job.ID, Status: app.RunnerJobExecutionStatusPending}
		if err := a.db.WithContext(ctx).Create(&execution).Error; err != nil {
			return nil, fmt.Errorf("unable to create runner job execution: %w", err)
		}
	}

	if err := a.updateJob(ctx, job.ID, app.RunnerJobStatusInProgress, "in-progress"); err != nil {
		return nil, err
	}
	return &EnsureExecutionResponse{ExecutionID: execution.ID, JobExecutionTimeout: job.ExecutionTimeout}, nil
}

type RunJobRequest struct {
	JobID       string `json:"job_id" validate:"required"`
	ExecutionID string `json:"execution_id" validate:"required"`
}

func (a *Activities) RunJob(ctx context.Context, req *RunJobRequest) error {
	activity.RecordHeartbeat(ctx, "starting")
	stopHeartbeat := runnercontrolplane.HeartbeatUntilDone(ctx, func() { activity.RecordHeartbeat(ctx, "running") })
	defer stopHeartbeat()

	job, err := a.getJob(ctx, req.JobID)
	if err != nil {
		return err
	}
	var execution app.RunnerJobExecution
	if err := a.db.WithContext(ctx).Where(&app.RunnerJobExecution{ID: req.ExecutionID, RunnerJobID: req.JobID}).First(&execution).Error; err != nil {
		return fmt.Errorf("unable to get runner job execution: %w", err)
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)

	executor, err := runnercontrolplane.NewExecutor(a, a.l, runnercontrolplane.Config{GitRef: a.cfg.GitRef})
	if err != nil {
		return fmt.Errorf("unable to create control-plane executor: %w", err)
	}
	return executor.Execute(ctx, toRunnerJobModel(job), toRunnerExecutionModel(&execution))
}

type FinalizeOutcome struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type FinalizeRequest struct {
	JobID       string          `json:"job_id" validate:"required"`
	ExecutionID string          `json:"execution_id" validate:"required"`
	Outcome     FinalizeOutcome `json:"outcome"`
}

func (a *Activities) Finalize(ctx context.Context, req *FinalizeRequest) error {
	jobStatus := app.RunnerJobStatusFinished
	execStatus := app.RunnerJobExecutionStatusFinished
	description := "finished"
	if !req.Outcome.Success {
		jobStatus = app.RunnerJobStatusFailed
		execStatus = app.RunnerJobExecutionStatusFailed
		description = req.Outcome.Error
		if err := a.ensureFailureResult(ctx, req.JobID, req.ExecutionID, req.Outcome.Error); err != nil {
			return err
		}
	}
	if _, err := a.UpdateJobExecution(ctx, req.JobID, req.ExecutionID, &models.ServiceUpdateRunnerJobExecutionRequest{Status: models.AppRunnerJobExecutionStatus(execStatus), StatusDescription: description}); err != nil {
		return err
	}
	return a.updateJob(ctx, req.JobID, jobStatus, description)
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
	updates := app.RunnerJobExecution{Status: app.RunnerJobExecutionStatus(req.Status)}
	if req.StatusDescription != "" {
		updates.StatusV2 = app.NewCompositeStatus(ctx, app.Status(req.Status))
		updates.StatusV2.StatusHumanDescription = truncateStatusDescription(req.StatusDescription)
	}
	var execution app.RunnerJobExecution
	if err := a.db.WithContext(ctx).Model(&execution).Where(&app.RunnerJobExecution{ID: executionID, RunnerJobID: jobID}).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("unable to update runner job execution: %w", err)
	}
	return toRunnerExecutionModel(&execution), nil
}

func (a *Activities) CreateJobExecutionResult(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)

	result := app.RunnerJobExecutionResult{OrgID: job.OrgID, RunnerJobExecutionID: executionID, Success: req.Success, Contents: req.Contents, ErrorCode: int(req.ErrorCode), ErrorMetadata: stringMapToHstore(req.ErrorMetadata)}
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

	var existing app.RunnerJobExecutionResult
	res := a.db.WithContext(ctx).Where(&app.RunnerJobExecutionResult{RunnerJobExecutionID: executionID}).First(&existing)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to get existing runner job execution result: %w", res.Error)
	}
	if res.Error == gorm.ErrRecordNotFound {
		if err := a.db.WithContext(ctx).Create(&result).Error; err != nil {
			return nil, fmt.Errorf("unable to create runner job execution result: %w", err)
		}
		return &models.AppRunnerJobExecutionResult{ID: result.ID, Success: result.Success}, nil
	}

	if err := a.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"org_id":                result.OrgID,
		"success":               result.Success,
		"error_code":            result.ErrorCode,
		"error_metadata":        result.ErrorMetadata,
		"contents":              result.Contents,
		"contents_gzip":         result.ContentsGzip,
		"contents_display":      result.ContentsDisplay,
		"contents_display_gzip": result.ContentsDisplayGzip,
	}).Error; err != nil {
		return nil, fmt.Errorf("unable to update runner job execution result: %w", err)
	}
	return &models.AppRunnerJobExecutionResult{ID: existing.ID, Success: result.Success}, nil
}

func (a *Activities) CreateJobExecutionOutputs(ctx context.Context, jobID, executionID string, req *models.ServiceCreateRunnerJobExecutionOutputsRequest) (*models.AppRunnerJobExecutionOutputs, error) {
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)
	byts, err := json.Marshal(req.Outputs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal outputs: %w", err)
	}
	outputs := app.RunnerJobExecutionOutputs{OrgID: job.OrgID, RunnerJobExecutionID: executionID, Outputs: byts}

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

	if err := a.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"org_id":  outputs.OrgID,
		"outputs": outputs.Outputs,
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
			"handler": "control-plane",
			"message": message,
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

func (a *Activities) updateJob(ctx context.Context, jobID string, status app.RunnerJobStatus, description string) error {
	job := app.RunnerJob{ID: jobID}
	if err := a.db.WithContext(ctx).Model(&job).Updates(app.RunnerJob{Status: status, StatusDescription: description}).Error; err != nil {
		return fmt.Errorf("unable to update runner job status: %w", err)
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
