package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
)

type AppListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// @ID						GetApps
// @Summary				get all apps for the current org
// @Description.markdown	get_apps.md
// @Param					offset						query	int		false	"offset of jobs to return"	Default(0)
// @Param					q							query	string	false	"search query to filter apps by name or ID"
// @Param					limit						query	int		false	"limit of jobs to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		AppListItem
// @Router					/v1/apps [get]
func (s *service) GetApps(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	limit, offset := 10, 0
	if page := cctx.OffsetPaginationFromContext(ctx); page != nil {
		limit, offset = page.Limit, page.Offset
	}
	apps, err := s.getCustomerApps(ctx.Request.Context(), org.ID, ctx.Query("q"), limit, offset)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for %s: %w", org.ID, err))
		return
	}
	apps, err = db.HandlePaginatedResponse(ctx, apps)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getCustomerApps(ctx context.Context, orgID, q string, limit, offset int) ([]AppListItem, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, stderr.NewInvalidRequest(fmt.Errorf("invalid app pagination"))
	}
	apps := []AppListItem{}
	err := authz.New(authz.Params{DB: s.db}).ReadCustomerResources(ctx, orgID, func(tx, readableApps, installs *gorm.DB) error {
		query := tx.Model(&app.App{}).Select("id", "name").Where(app.App{OrgID: orgID}).
			Where("id IN (?) OR id IN (?)", readableApps.Select("apps.id"), installs.Select("app_id")).
			Order("LOWER(name), id").Limit(limit + 1).Offset(offset)
		if q = strings.TrimSpace(q); q != "" {
			pattern := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(q) + "%"
			query = query.Where("name ILIKE ? OR id ILIKE ?", pattern, pattern)
		}
		return query.Find(&apps).Error
	})
	return apps, err
}
