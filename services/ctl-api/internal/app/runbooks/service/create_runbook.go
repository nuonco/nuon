package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/ensureinstallrunbooks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CreateRunbookRequest struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// @ID				CreateRunbook
// @Summary		create a runbook for an app
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id	path		string				true	"app ID"
// @Param			req		body		CreateRunbookRequest	true	"Input"
// @Success		201		{object}	app.Runbook
// @Failure		400		{object}	stderr.ErrResponse
// @Failure		401		{object}	stderr.ErrResponse
// @Failure		403		{object}	stderr.ErrResponse
// @Failure		404		{object}	stderr.ErrResponse
// @Failure		500		{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/runbooks [post]
func (s *service) CreateRunbook(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	appID := ctx.Param("app_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateRunbookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	runbook := app.Runbook{
		AppID:       appID,
		OrgID:       org.ID,
		Name:        req.Name,
		Description: req.Description,
	}
	if len(req.Labels) > 0 {
		runbook.Labels = labels.Labels(req.Labels)
	}

	res := s.db.WithContext(ctx).Create(&runbook)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create runbook: %w", res.Error))
		return
	}

	// fan the new runbook out across every install for the app durably, off the
	// request path, via a signal on the app queue
	queue, err := s.appsHelpers.GetOrCreateAppSignalsQueue(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app signals queue: %w", err))
		return
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   appID,
		OwnerType: "apps",
		Signal:    &ensureinstallrunbooks.Signal{AppID: appID},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue ensure install runbooks signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, runbook)
}
