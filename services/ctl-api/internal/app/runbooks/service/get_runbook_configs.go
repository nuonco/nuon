package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID				GetRunbookConfigs
// @Summary		get runbook configs
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id		path	string	true	"app ID"
// @Param			runbook_id	path	string	true	"runbook ID"
// @Param			offset		query	int		false	"offset"	Default(0)
// @Param			limit		query	int		false	"limit"		Default(10)
// @Success		200			{array}	app.RunbookConfig
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id}/configs [get]
func (s *service) GetRunbookConfigs(ctx *gin.Context) {
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

	configs := []*app.RunbookConfig{}
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.RunbookConfig{OrgID: org.ID, RunbookID: runbookID}).
		Order("created_at DESC").
		Find(&configs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbook configs: %w", res.Error))
		return
	}

	configs, err = db.HandlePaginatedResponse(ctx, configs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}
