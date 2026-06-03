package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID				GetRunbooks
// @Summary		get runbooks for an app
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id	path	string	true	"app ID"
// @Param			q		query	string	false	"search query to filter runbooks by name or ID"
// @Param			offset	query	int		false	"offset"	Default(0)
// @Param			limit	query	int		false	"limit"		Default(10)
// @Success		200		{array}	app.Runbook
// @Router			/v1/apps/{app_id}/runbooks [get]
func (s *service) GetRunbooks(ctx *gin.Context) {
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

	q := ctx.Query("q")
	lbls := labels.ParseLabelsQuery(ctx.Query("labels"))

	runbooks := []*app.Runbook{}
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Scopes(labels.WithLabels("labels", lbls)).
		Preload("Configs", func(tx2 *gorm.DB) *gorm.DB {
			return tx2.Scopes(scopes.WithOverrideTable("runbook_configs_latest_view_v1"))
		}).
		Preload("Configs.Steps", func(tx2 *gorm.DB) *gorm.DB {
			return tx2.Order("idx ASC")
		}).
		Where(app.Runbook{OrgID: org.ID, AppID: appID})

	if q != "" {
		tx = tx.Where("name ILIKE ? OR id = ?", "%"+q+"%", q)
	}

	res := tx.Find(&runbooks)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbooks: %w", res.Error))
		return
	}

	runbooks, err = db.HandlePaginatedResponse(ctx, runbooks)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runbooks)
}
