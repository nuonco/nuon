package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sandboxbuildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/sandboxbuild"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type CreateAppSandboxBuildRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
}

// @ID        CreateAppSandboxBuild
// @Summary   create app sandbox build
// @Tags      apps
// @Accept    json
// @Produce   json
// @Param     app_id  path  string  true  "app ID"
// @Param     req     body  CreateAppSandboxBuildRequest  false  "Input"
// @Security  APIKey
// @Security  OrgID
// @Failure   400  {object}  stderr.ErrResponse
// @Failure   401  {object}  stderr.ErrResponse
// @Failure   404  {object}  stderr.ErrResponse
// @Failure   409  {object}  stderr.ErrResponse
// @Failure   500  {object}  stderr.ErrResponse
// @Success   201  {object}  app.AppSandboxBuild
// @Router    /v1/apps/{app_id}/sandbox/builds [post]
func (s *service) CreateAppSandboxBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppSandboxBuildRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := request.ValidateRequestID(req.RequestID); err != nil {
		ctx.Error(err)
		return
	}

	currentApp, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	var requestHash string
	if req.RequestID != "" {
		requestHash, err = request.Hash(struct {
			AppID string `json:"app_id"`
		}{AppID: currentApp.ID})
		if err != nil {
			ctx.Error(err)
			return
		}
		existing, lookupErr := s.lookupAppSandboxBuild(ctx, currentApp.OrgID, currentApp.ID, req.RequestID)
		if lookupErr == nil {
			if err := request.Check(existing.Request, requestHash, ""); err != nil {
				ctx.Error(err)
				return
			}
			if err := s.enqueueAppSandboxBuild(ctx, currentApp.ID, existing, req.RequestID); err != nil {
				ctx.Error(err)
				return
			}
			ctx.JSON(http.StatusCreated, existing)
			return
		}
		if !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			ctx.Error(fmt.Errorf("unable to look up sandbox build: %w", lookupErr))
			return
		}
	}

	// Get the latest app config
	var latestConfig app.AppConfig
	if res := s.db.WithContext(ctx).
		Where("app_id = ?", currentApp.ID).
		Order("created_at DESC").
		First(&latestConfig); res.Error != nil {
		ctx.Error(fmt.Errorf("no app config found for app %s: %w", appID, res.Error))
		return
	}

	// Get the latest sandbox config
	if len(currentApp.AppSandboxConfigs) == 0 {
		ctx.Error(fmt.Errorf("no sandbox config found for app %s", appID))
		return
	}
	latestSandboxConfig := currentApp.AppSandboxConfigs[0]

	// Create the build record immediately so the caller gets an ID back
	var wfRequest *app.WorkflowRequest
	if req.RequestID != "" {
		wfRequest = &app.WorkflowRequest{
			RequestID:         req.RequestID,
			RequestHash:       requestHash,
			PinnedAppConfigID: "",
		}
	}
	build := app.AppSandboxBuild{
		AppID:              currentApp.ID,
		AppConfigID:        latestConfig.ID,
		AppSandboxConfigID: latestSandboxConfig.ID,
		OrgID:              currentApp.OrgID,
		Status:             app.AppSandboxBuildStatusQueued,
		StatusDescription:  "queued and waiting for runner",
		Request:            wfRequest,
	}
	if res := s.db.WithContext(ctx).Create(&build); res.Error != nil {
		if req.RequestID != "" && request.IsDuplicateKey(res.Error) {
			existing, lookupErr := s.lookupAppSandboxBuild(ctx, currentApp.OrgID, currentApp.ID, req.RequestID)
			if lookupErr != nil {
				ctx.Error(fmt.Errorf("unable to look up sandbox build: %w", lookupErr))
				return
			}
			if err := request.Check(existing.Request, requestHash, ""); err != nil {
				ctx.Error(err)
				return
			}
			if err := s.enqueueAppSandboxBuild(ctx, currentApp.ID, existing, req.RequestID); err != nil {
				ctx.Error(err)
				return
			}
			ctx.JSON(http.StatusCreated, existing)
			return
		}
		ctx.Error(fmt.Errorf("unable to create sandbox build: %w", res.Error))
		return
	}

	if err := s.enqueueAppSandboxBuild(ctx, currentApp.ID, &build, req.RequestID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, build)
}

func (s *service) lookupAppSandboxBuild(ctx *gin.Context, orgID, appID, requestID string) (*app.AppSandboxBuild, error) {
	var existing app.AppSandboxBuild
	err := s.db.WithContext(ctx).
		Where(app.AppSandboxBuild{OrgID: orgID, AppID: appID}).
		Where("request->>'request_id' = ?", requestID).
		First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *service) enqueueAppSandboxBuild(ctx *gin.Context, appID string, build *app.AppSandboxBuild, requestID string) error {
	q, err := s.queueClient.GetDefaultQueueByOwner(ctx, appID, "apps")
	if err != nil {
		return fmt.Errorf("unable to get sandbox queue for app %s: %w", appID, err)
	}
	enq := &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &sandboxbuildsignal.Signal{
			AppConfigID:       build.AppConfigID,
			AppSandboxBuildID: build.ID,
		},
	}
	if requestID != "" {
		dedupe := request.DedupeKey("sandbox-build", requestID)
		enq.DedupeKey = &dedupe
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, enq); err != nil {
		return fmt.Errorf("unable to enqueue sandbox build signal: %w", err)
	}
	return nil
}
