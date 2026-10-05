package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type IdempotentInstallWorkflowRequest struct {
	InstallID             string
	WorkflowType          app.WorkflowType
	Metadata              map[string]string
	PlanOnly              bool
	Role                  string
	RequestID             string
	RequestHash           string
	Operation             string
	QueueName             string
	SkipAppConfigConflict bool
}

type IdempotentInstallWorkflowHooks struct {
	Prepare     func(tx *gorm.DB) (map[string]string, error)
	AfterCreate func(tx *gorm.DB, wf *app.Workflow) error
}

func (h *Helpers) RunIdempotentInstallWorkflow(ctx context.Context, req IdempotentInstallWorkflowRequest, hooks *IdempotentInstallWorkflowHooks) (*app.Workflow, bool, error) {
	if err := request.ValidateRequestID(req.RequestID); err != nil {
		return nil, false, err
	}
	if req.RequestID == "" {
		return nil, false, fmt.Errorf("request_id is required")
	}
	if req.RequestHash == "" {
		return nil, false, fmt.Errorf("request hash is required")
	}

	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "org_id", "app_config_id", "name").
		Where(app.Install{ID: req.InstallID}).
		First(&install).Error; err != nil {
		return nil, false, fmt.Errorf("unable to get install: %w", err)
	}

	var wf *app.Workflow
	var created bool
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		wf, created, err = h.attemptIdempotentInstallWorkflow(ctx, &install, req, hooks)
		if err == nil || !request.IsDuplicateKey(err) {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 20 * time.Millisecond)
		var current app.Install
		if reloadErr := h.db.WithContext(ctx).Select("app_config_id").Where(app.Install{ID: install.ID}).First(&current).Error; reloadErr != nil {
			return nil, false, fmt.Errorf("unable to get install: %w", reloadErr)
		}
		install.AppConfigID = current.AppConfigID
		found, lookupErr := h.lookupIdempotentWorkflow(ctx, h.db, &install, req)
		if lookupErr == nil {
			if err = checkIdempotentWorkflow(found.Request, req, install.AppConfigID); err != nil {
				return nil, false, err
			}
			wf = found
			created = false
			err = nil
			break
		}
		if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			return nil, false, lookupErr
		}
	}
	if err != nil {
		if request.IsDuplicateKey(err) {
			return nil, false, stderr.ErrConflict{
				Err:         fmt.Errorf("request_id is already in use"),
				Description: "request_id is already in use",
			}
		}
		return nil, false, err
	}

	h.wakeIdempotentInstallWorkflow(ctx, install.ID, req.QueueName, wf.ID, request.DedupeKey(req.Operation, req.RequestID))
	return wf, created, nil
}

func (h *Helpers) attemptIdempotentInstallWorkflow(ctx context.Context, install *app.Install, req IdempotentInstallWorkflowRequest, hooks *IdempotentInstallWorkflowHooks) (*app.Workflow, bool, error) {
	var wf *app.Workflow
	var created bool
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current app.Install
		if err := tx.Select("id", "app_config_id").Where(app.Install{ID: install.ID}).First(&current).Error; err != nil {
			return fmt.Errorf("unable to get install: %w", err)
		}
		install.AppConfigID = current.AppConfigID

		existing, err := h.lookupIdempotentWorkflow(ctx, tx, install, req)
		if err == nil {
			if err := checkIdempotentWorkflow(existing.Request, req, install.AppConfigID); err != nil {
				return err
			}
			wf = existing
			created = false
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		md := map[string]string{}
		for k, v := range req.Metadata {
			md[k] = v
		}
		if hooks != nil && hooks.Prepare != nil {
			extra, err := hooks.Prepare(tx)
			if err != nil {
				return err
			}
			for k, v := range extra {
				md[k] = v
			}
		}

		createdWorkflow, err := h.insertInstallWorkflow(ctx, tx, install.ID, req.WorkflowType, md, req.PlanOnly, req.Role, &app.WorkflowRequest{
			RequestID:         req.RequestID,
			RequestHash:       req.RequestHash,
			PinnedAppConfigID: install.AppConfigID,
		})
		if err != nil {
			return err
		}
		if hooks != nil && hooks.AfterCreate != nil {
			if err := hooks.AfterCreate(tx, createdWorkflow); err != nil {
				return err
			}
		}

		var q app.Queue
		if err := tx.Where(app.Queue{OwnerID: install.ID, Name: req.QueueName}).First(&q).Error; err != nil {
			return fmt.Errorf("unable to find %s queue for install %s: %w", req.QueueName, install.ID, err)
		}
		dedupe := request.DedupeKey(req.Operation, req.RequestID)
		if _, err := h.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			Signal:    executeflow.NewSignal(createdWorkflow.ID),
			OwnerID:   createdWorkflow.ID,
			OwnerType: "install_workflows",
			DedupeKey: &dedupe,
		}); err != nil {
			return err
		}
		wf = createdWorkflow
		created = true
		return nil
	})
	return wf, created, err
}

func checkIdempotentWorkflow(idem *app.WorkflowRequest, req IdempotentInstallWorkflowRequest, currentAppConfigID string) error {
	if req.SkipAppConfigConflict && idem != nil {
		currentAppConfigID = idem.PinnedAppConfigID
	}
	return request.Check(idem, req.RequestHash, currentAppConfigID)
}

func (h *Helpers) lookupIdempotentWorkflow(ctx context.Context, db *gorm.DB, install *app.Install, req IdempotentInstallWorkflowRequest) (*app.Workflow, error) {
	var existing app.Workflow
	err := db.WithContext(ctx).
		Where(app.Workflow{
			OrgID:     install.OrgID,
			OwnerID:   install.ID,
			OwnerType: "installs",
			Type:      req.WorkflowType,
		}).
		Where("request->>'request_id' = ?", req.RequestID).
		First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func (h *Helpers) wakeIdempotentInstallWorkflow(ctx context.Context, installID, queueName, workflowID, dedupe string) {
	var q app.Queue
	if err := h.db.WithContext(ctx).Where(app.Queue{OwnerID: installID, Name: queueName}).First(&q).Error; err != nil {
		cctx.GetLogger(ctx, h.l).Warn("idempotent workflow signal wake failed",
			zap.Error(err),
			zap.String("workflow_id", workflowID),
		)
		return
	}
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		Signal:    executeflow.NewSignal(workflowID),
		OwnerID:   workflowID,
		OwnerType: "install_workflows",
		DedupeKey: &dedupe,
	}); err != nil {
		cctx.GetLogger(ctx, h.l).Warn("idempotent workflow signal wake failed",
			zap.Error(err),
			zap.String("workflow_id", workflowID),
		)
	}
}
