package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type UpdateRunbookRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// @ID				UpdateRunbook
// @Summary		update a runbook
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id		path	string					true	"app ID"
// @Param			runbook_id	path	string					true	"runbook ID"
// @Param			req			body	UpdateRunbookRequest	true	"Input"
// @Success		200			{object}	app.Runbook
// @Failure		400		{object}	stderr.ErrResponse
// @Failure		401		{object}	stderr.ErrResponse
// @Failure		403		{object}	stderr.ErrResponse
// @Failure		404		{object}	stderr.ErrResponse
// @Failure		500		{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id} [patch]
func (s *service) UpdateRunbook(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	runbookID := ctx.Param("runbook_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var runbook app.Runbook
	if res := s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: org.ID}).
		Where("id = ?", runbookID).
		First(&runbook); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbook: %w", res.Error))
		return
	}

	var req UpdateRunbookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if req.Name != "" {
		runbook.Name = req.Name
	}
	if req.Description != "" {
		runbook.Description = req.Description
	}
	if len(req.Labels) > 0 {
		runbook.Labels = labels.Labels(req.Labels)
	}

	if res := s.db.WithContext(ctx).Save(&runbook); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update runbook: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, runbook)
}
