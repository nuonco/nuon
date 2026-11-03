package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
func (s *service) BuildAppConfig(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}
	target, tx := access.Target(), access.DB()
	opCtx := tx.Statement.Context
	var req BuildAppConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		return require.Response{}, stderr.NewInvalidRequest(err)
	}
	if err := request.ValidateRequestID(req.RequestID); err != nil {
		return require.Response{}, err
	}
	configID := ctx.Param("config_id")
	if configID == "" {
		return require.Response{}, stderr.NewInvalidRequest(fmt.Errorf("config ID is required"))
	}
	var config app.AppConfig
	if err := tx.Where(app.AppConfig{ID: configID, AppID: target.ID, OrgID: target.OrgID}).Take(&config).Error; err != nil {
		return require.Response{}, fmt.Errorf("unable to find app config: %w", err)
	}
	requestHash, err := request.Hash(struct {
		AppID    string `json:"app_id"`
		ConfigID string `json:"config_id"`
	}{target.ID, config.ID})
	if err != nil {
		return require.Response{}, err
	}
	if req.RequestID != "" {
		var existing app.Workflow
		err := tx.Where(app.Workflow{OrgID: target.OrgID, OwnerID: target.ID, OwnerType: "apps", Type: app.WorkflowTypeAppConfigBuild}).
			Where("request->>'request_id' = ?", req.RequestID).Take(&existing).Error
		if err == nil {
			if err := request.Check(existing.Request, requestHash, config.ID); err != nil {
				return require.Response{}, err
			}
			credential, _ := cctx.VerifiedCredentialFromContext(opCtx)
			if existing.CreatedByID != credential.AccountID || len(existing.Authorization) == 0 || existing.AuthorizationFailure != "" {
				return require.Response{}, stderr.ErrConflict{Err: fmt.Errorf("request cannot resume this operation"), Description: "submit a new request ID"}
			}
			return require.Response{Status: http.StatusCreated, Body: &existing}, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return require.Response{}, err
		}
	}
	var queue app.Queue
	if err := tx.Where(app.Queue{OwnerID: target.ID, OwnerType: "apps", Name: appshelpers.AppWorkflowsQueueName}).Take(&queue).Error; err != nil {
		return require.Response{}, fmt.Errorf("app-workflows queue must be provisioned before accepting work: %w", err)
	}
	wf := appshelpers.NewWorkflow(opCtx, target.ID, "apps", app.WorkflowTypeAppConfigBuild,
		map[string]string{"app_config_id": config.ID}, false, app.InstallApprovalOptionPrompt, "", &app.WorkflowRequest{
			RequestID: req.RequestID, RequestHash: requestHash, PinnedAppConfigID: config.ID,
		})
	if err := access.AcceptWorkflow(wf, time.Now().Add(callback.MaxWaitCeiling)); err != nil {
		return require.Response{}, err
	}
	if _, err := s.queueClient.EnqueueSignalInTransaction(opCtx, tx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID, Signal: executeflow.NewSignal(wf.ID), OwnerID: wf.ID, OwnerType: "install_workflows",
	}); err != nil {
		return require.Response{}, fmt.Errorf("unable to enqueue build signal: %w", err)
	}
	return require.Response{Status: http.StatusCreated, Body: wf}, nil
}
