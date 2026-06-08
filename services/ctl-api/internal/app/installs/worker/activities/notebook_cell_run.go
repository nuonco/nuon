package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateNotebookCellRunRequest struct {
	NotebookID     string `validate:"required"`
	CellID         string `validate:"required"`
	IdempotencyKey string `validate:"required"`
	// OrgID / InstallID / TriggeredByID are passed explicitly because the
	// long-lived notebook workflow is started once and serves runs from many
	// accounts — the workflow's start context cannot be trusted for per-run
	// ownership/audit fields.
	OrgID         string `validate:"required"`
	InstallID     string `validate:"required"`
	TriggeredByID string `validate:"required"`
}

type CreateNotebookCellRunResponse struct {
	NotebookCellRunID          string
	InstallActionWorkflowRunID string
	RunnerID                   string
	Role                       string
	// AlreadyDispatched is true when an existing run for this idempotency key
	// already has an action run, so the caller must not enqueue it again.
	AlreadyDispatched bool
}

// CreateNotebookCellRun idempotently (keyed on notebook_id + idempotency_key)
// creates a NotebookCellRun plus the underlying adhoc-shaped
// InstallActionWorkflowRun that the existing dispatch path operates on. The
// cell config is snapshotted onto both records so history stays truthful even
// after the cell is later edited or deleted.
//
// @temporal-gen-v2 activity
func (a *Activities) CreateNotebookCellRun(ctx context.Context, req *CreateNotebookCellRunRequest) (*CreateNotebookCellRunResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	var cell app.NotebookCell
	if res := a.db.WithContext(ctx).
		Preload("Notebook").
		First(&cell, "id = ? AND notebook_id = ? AND org_id = ?", req.CellID, req.NotebookID, req.OrgID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get notebook cell")
	}
	if cell.Notebook.InstallID != req.InstallID {
		return nil, errors.New("notebook does not belong to install")
	}
	if cell.Notebook.Status == app.NotebookStatusArchived {
		return nil, errors.New("notebook is archived")
	}

	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	// Idempotency: if a run already exists for this key, return it as-is.
	var existing app.NotebookCellRun
	res := a.db.WithContext(ctx).
		Where(app.NotebookCellRun{NotebookID: req.NotebookID, IdempotencyKey: req.IdempotencyKey}).
		First(&existing)
	if res.Error == nil {
		return &CreateNotebookCellRunResponse{
			NotebookCellRunID:          existing.ID,
			InstallActionWorkflowRunID: existing.InstallActionWorkflowRunID,
			RunnerID:                   install.RunnerID,
			Role:                       cell.Role,
			AlreadyDispatched:          existing.InstallActionWorkflowRunID != "",
		}, nil
	}

	// Build the adhoc-shaped action run from the cell snapshot. This reuses the
	// proven adhoc plan + dispatch path; notebook runs are identified by the
	// NotebookCellRun row, not by a distinct trigger type.
	stepConfig := app.ActionWorkflowStepConfig{
		InlineContents: cell.InlineContents,
		Command:        cell.Command,
		EnvVars:        cell.EnvVars,
		Name:           cellRunName(&cell),
		Idx:            0,
	}
	adHocConfig := app.AdHocStepConfig(stepConfig)

	run := app.InstallActionWorkflowRun{
		OrgID:             req.OrgID,
		CreatedByID:       req.TriggeredByID,
		InstallID:         install.ID,
		TriggerType:       app.ActionWorkflowTriggerTypeAdHoc,
		TriggeredByID:     req.TriggeredByID,
		TriggeredByType:   "account",
		Status:            app.InstallActionRunStatusQueued,
		StatusDescription: "Queued for execution",
		Steps: []app.InstallActionWorkflowRunStep{{
			Status:      app.InstallActionWorkflowRunStepStatusPending,
			AdHocConfig: &adHocConfig,
		}},
		RunEnvVars:       cell.EnvVars,
		Timeout:          cell.Timeout,
		Role:             cell.Role,
		EnableKubeConfig: cell.EnableKubeConfig,
	}
	if err := a.db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, errors.Wrap(err, "unable to create action workflow run")
	}

	cellRun := app.NotebookCellRun{
		OrgID:                      req.OrgID,
		CreatedByID:                req.TriggeredByID,
		InstallID:                  install.ID,
		NotebookID:                 req.NotebookID,
		CellID:                     req.CellID,
		CellRevision:               cell.Revision,
		IdempotencyKey:             req.IdempotencyKey,
		InstallActionWorkflowRunID: run.ID,
		Name:                       cell.Name,
		InlineContents:             cell.InlineContents,
		Command:                    cell.Command,
		EnvVars:                    cell.EnvVars,
		TriggeredByID:              req.TriggeredByID,
		TriggeredByType:            "account",
		Status:                     app.InstallActionRunStatusQueued,
		StatusDescription:          "Queued for execution",
	}
	if err := a.db.WithContext(ctx).Create(&cellRun).Error; err != nil {
		return nil, errors.Wrap(err, "unable to create notebook cell run")
	}

	return &CreateNotebookCellRunResponse{
		NotebookCellRunID:          cellRun.ID,
		InstallActionWorkflowRunID: run.ID,
		RunnerID:                   install.RunnerID,
		Role:                       cell.Role,
	}, nil
}

type UpdateNotebookCellRunRequest struct {
	NotebookCellRunID string `validate:"required"`
	Status            app.InstallActionWorkflowRunStatus
	StatusDescription string
	LogStreamID       string
	RunnerJobID       string
}

// UpdateNotebookCellRun updates a NotebookCellRun's status and (optionally) its
// dispatch artifact IDs (log stream / runner job) as the run progresses, so the
// UI can tail logs and reflect status from the cell run row.
//
// @temporal-gen-v2 activity
func (a *Activities) UpdateNotebookCellRun(ctx context.Context, req *UpdateNotebookCellRunRequest) error {
	if err := a.v.Struct(req); err != nil {
		return errors.Wrap(err, "invalid request")
	}

	updates := map[string]any{}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.StatusDescription != "" {
		updates["status_description"] = req.StatusDescription
	}
	if req.LogStreamID != "" {
		updates["log_stream_id"] = req.LogStreamID
	}
	if req.RunnerJobID != "" {
		updates["runner_job_id"] = req.RunnerJobID
	}
	if len(updates) == 0 {
		return nil
	}

	res := a.db.WithContext(ctx).
		Model(&app.NotebookCellRun{}).
		Where("id = ?", req.NotebookCellRunID).
		Updates(updates)
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update notebook cell run")
	}
	return nil
}

func cellRunName(cell *app.NotebookCell) string {
	if cell.Name != "" {
		return cell.Name
	}
	if cell.InlineContents != "" {
		return "Notebook script"
	}
	return "Notebook command"
}
