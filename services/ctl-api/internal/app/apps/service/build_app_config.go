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
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type BuildAppConfigRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
}

// @ID						BuildAppConfig
// @Summary				Build all components for an app config
// @Description			Creates a workflow that builds all components defined in the given app config.
// @Tags					apps
// @Accept					json
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
// @Param					req			body	BuildAppConfigRequest	false	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Workflow
// @Router					/v1/apps/{app_id}/configs/{config_id}/build [post]
func (s *service) BuildAppConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	var req BuildAppConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := request.ValidateRequestID(req.RequestID); err != nil {
		ctx.Error(err)
		return
	}

	// Verify app exists and belongs to this org
	var a app.App
	res := s.db.WithContext(ctx).
		Where(app.App{OrgID: org.ID}).
		First(&a, "id = ?", appID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app: %w", res.Error))
		return
	}

	// Verify config exists and belongs to this app
	var config app.AppConfig
	res = s.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID, OrgID: org.ID}).
		First(&config, "id = ?", configID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app config: %w", res.Error))
		return
	}

	// Ensure the app has a queue for signal routing
	if err := s.helpers.EnsureAppQueue(ctx, appID); err != nil {
		ctx.Error(fmt.Errorf("unable to ensure app queue: %w", err))
		return
	}

	queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, appID, plugins.TableName(s.db, app.App{}), appshelpers.AppWorkflowsQueueName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find app-workflows queue: %w", err))
		return
	}

	var requestHash string
	if req.RequestID != "" {
		requestHash, err = request.Hash(struct {
			AppID    string `json:"app_id"`
			ConfigID string `json:"config_id"`
		}{AppID: appID, ConfigID: configID})
		if err != nil {
			ctx.Error(err)
			return
		}
		existing, lookupErr := s.lookupAppConfigBuild(ctx, org.ID, appID, req.RequestID)
		if lookupErr == nil {
			if err := request.Check(existing.Request, requestHash, ""); err != nil {
				ctx.Error(err)
				return
			}
			if err := s.enqueueAppConfigBuild(ctx, queue.ID, existing.ID, req.RequestID); err != nil {
				ctx.Error(err)
				return
			}
			ctx.JSON(http.StatusCreated, existing)
			return
		}
		if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			ctx.Error(fmt.Errorf("unable to look up app config build: %w", lookupErr))
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
	wf, err := s.helpers.CreateAppWorkflowWithRequest(
		ctx,
		appID,
		app.WorkflowTypeAppConfigBuild,
		map[string]string{
			"app_config_id": configID,
		},
		false,
		wfRequest,
	)
	if err != nil && req.RequestID != "" && request.IsDuplicateKey(err) {
		existing, lookupErr := s.lookupAppConfigBuild(ctx, org.ID, appID, req.RequestID)
		if lookupErr != nil {
			ctx.Error(fmt.Errorf("unable to look up app config build: %w", lookupErr))
			return
		}
		if err := request.Check(existing.Request, requestHash, ""); err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueAppConfigBuild(ctx, queue.ID, existing.ID, req.RequestID); err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusCreated, existing)
		return
	}
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	if err := s.enqueueAppConfigBuild(ctx, queue.ID, wf.ID, req.RequestID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, wf)
}

func (s *service) lookupAppConfigBuild(ctx *gin.Context, orgID, appID, requestID string) (*app.Workflow, error) {
	var existing app.Workflow
	err := s.db.WithContext(ctx).
		Where(app.Workflow{
			OrgID:     orgID,
			OwnerID:   appID,
			OwnerType: "apps",
			Type:      app.WorkflowTypeAppConfigBuild,
		}).
		Where("request->>'request_id' = ?", requestID).
		First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *service) enqueueAppConfigBuild(ctx *gin.Context, queueID, workflowID, requestID string) error {
	enq := &queueclient.EnqueueSignalRequest{
		QueueID: queueID,
		Signal:  executeflow.NewSignal(workflowID),
	}
	if requestID != "" {
		dedupe := request.DedupeKey("app-config-build", requestID)
		enq.DedupeKey = &dedupe
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, enq); err != nil {
		return fmt.Errorf("unable to enqueue build signal: %w", err)
	}
	return nil
}
