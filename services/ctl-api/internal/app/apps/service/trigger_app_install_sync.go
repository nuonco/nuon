package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/syncinstalls"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type TriggerAppInstallSyncRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
}

// @ID						TriggerAppInstallSync
// @Summary				trigger app-level install config sync
// @Description			Triggers a sync of all install configs for the app from the configured git source.
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					app_id	path	string	true	"app ID"
// @Param					req		body	TriggerAppInstallSyncRequest	false	"Input"
// @Success				202	{object}	app.AppInstallConfigSync
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/install-syncs [post]
func (s *service) TriggerAppInstallSync(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")

	var req TriggerAppInstallSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := request.ValidateRequestID(req.RequestID); err != nil {
		ctx.Error(err)
		return
	}

	var a app.App
	if err := s.db.WithContext(ctx).
		Where(app.App{ID: appID, OrgID: org.ID}).
		First(&a).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	var latestConfig app.AppInstallsConfig
	if err := s.db.WithContext(ctx).
		Where(app.AppInstallsConfig{AppID: appID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		ctx.Error(fmt.Errorf("no installs config found for app - configure via installs.toml or the dashboard"))
		return
	}

	if err := s.helpers.EnsureAppQueue(ctx, appID); err != nil {
		ctx.Error(fmt.Errorf("unable to ensure app queues: %w", err))
		return
	}

	var requestHash string
	if req.RequestID != "" {
		requestHash, err = request.Hash(struct {
			AppID string `json:"app_id"`
		}{AppID: appID})
		if err != nil {
			ctx.Error(err)
			return
		}
		existing, lookupErr := s.lookupAppInstallConfigSync(ctx, org.ID, appID, req.RequestID)
		if lookupErr == nil {
			if err := request.Check(existing.Request, requestHash, ""); err != nil {
				ctx.Error(err)
				return
			}
			if _, _, err := s.enqueueAppInstallSync(ctx, appID, existing, account.ID, req.RequestID); err != nil {
				ctx.Error(err)
				return
			}
			ctx.JSON(http.StatusAccepted, existing)
			return
		}
		if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			ctx.Error(fmt.Errorf("unable to look up install config sync: %w", lookupErr))
			return
		}
	}

	var wfRequest *app.WorkflowRequest
	if req.RequestID != "" {
		wfRequest = &app.WorkflowRequest{
			RequestID:         req.RequestID,
			RequestHash:       requestHash,
			PinnedAppConfigID: "",
		}
	}
	syncRecord := app.AppInstallConfigSync{
		AppID:       appID,
		TriggeredBy: "manual",
		Status: app.CompositeStatus{
			Status: app.StatusQueued,
		},
		Request: wfRequest,
	}
	if err := s.db.WithContext(ctx).Create(&syncRecord).Error; err != nil {
		if req.RequestID != "" && request.IsDuplicateKey(err) {
			existing, lookupErr := s.lookupAppInstallConfigSync(ctx, org.ID, appID, req.RequestID)
			if lookupErr != nil {
				ctx.Error(fmt.Errorf("unable to look up install config sync: %w", lookupErr))
				return
			}
			if err := request.Check(existing.Request, requestHash, ""); err != nil {
				ctx.Error(err)
				return
			}
			if _, _, err := s.enqueueAppInstallSync(ctx, appID, existing, account.ID, req.RequestID); err != nil {
				ctx.Error(err)
				return
			}
			ctx.JSON(http.StatusAccepted, existing)
			return
		}
		ctx.Error(fmt.Errorf("unable to create install config sync: %w", err))
		return
	}

	enqueueResp, queue, err := s.enqueueAppInstallSync(ctx, appID, &syncRecord, account.ID, req.RequestID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.db.WithContext(ctx).Model(&syncRecord).Updates(map[string]any{
		"queue_signal_id": enqueueResp.ID,
		"queue_id":        queue.ID,
	})
	syncRecord.QueueSignalID = enqueueResp.ID
	syncRecord.QueueID = queue.ID

	ctx.JSON(http.StatusAccepted, syncRecord)
}

func (s *service) lookupAppInstallConfigSync(ctx *gin.Context, orgID, appID, requestID string) (*app.AppInstallConfigSync, error) {
	var existing app.AppInstallConfigSync
	err := s.db.WithContext(ctx).
		Where(app.AppInstallConfigSync{OrgID: orgID, AppID: appID}).
		Where("request->>'request_id' = ?", requestID).
		First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *service) enqueueAppInstallSync(ctx *gin.Context, appID string, syncRecord *app.AppInstallConfigSync, accountID, requestID string) (*queue.EnqueueResponse, *app.Queue, error) {
	queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, appID, "apps", appshelpers.AppInstallSyncsQueueName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find install-syncs queue for app: %w", err)
	}
	enq := &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   syncRecord.ID,
		OwnerType: "app_install_config_syncs",
		Signal: &syncinstalls.Signal{
			AppID:                  appID,
			AppInstallConfigSyncID: syncRecord.ID,
			TriggeredBy:            "manual",
			FallbackCreatedByID:    accountID,
		},
	}
	if requestID != "" {
		dedupe := request.DedupeKey("app-install-sync", requestID)
		enq.DedupeKey = &dedupe
	}
	resp, err := s.queueClient.EnqueueSignal(ctx, enq)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to enqueue sync signal: %w", err)
	}
	return resp, queue, nil
}
