package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type AdminAppDetails struct {
	*app.App

	Components []AdminAppComponentDetails `json:"components,omitempty"`
}

type AdminAppComponentDetails struct {
	*app.Component
}

const adminAppDetailsDefaultLimit = 25

// @ID			AdminListAppsDetails
// @BasePath	/v1/apps
// @Summary	Return a compact admin list of apps with their components and latest build status
// @Description	Admin list of apps intended for status / README rollups.
// @Description	Each app includes its components and each component's most recent build status.
// @Description	Pagination is uncapped on this admin endpoint — pass any `limit`.
// @Description	The optional `status` query parameter filters by
// @Description	`status_v2->>'status'` and may be repeated.
// @Param			offset	query	int			false	"offset of results to return"	Default(0)
// @Param			limit	query	int			false	"limit of results to return (no upper cap)"	Default(25)
// @Param			status	query	[]string	false	"filter by composite status (repeatable)"	collectionFormat(multi)
// @Tags			apps/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array}	AdminAppDetails
// @Router			/v1/apps/details [GET]
func (s *service) AdminListAppsDetails(ctx *gin.Context) {
	limit, offset := parseAdminAppDetailsPagination(ctx)
	statuses := ctx.QueryArray("status")

	items, err := s.listAppsDetails(ctx, limit, offset, statuses)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, items)
}

func parseAdminAppDetailsPagination(ctx *gin.Context) (limit, offset int) {
	limit = adminAppDetailsDefaultLimit
	if v, err := strconv.Atoi(ctx.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(ctx.Query("offset")); err == nil && v >= 0 {
		offset = v
	}
	return limit, offset
}

func (s *service) listAppsDetails(ctx *gin.Context, limit, offset int, statuses []string) ([]*AdminAppDetails, error) {
	var apps []*app.App
	tx := s.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Offset(offset)
	if len(statuses) > 0 {
		tx = tx.Where("status_v2->>'status' IN ?", statuses)
	}
	if err := tx.Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("unable to list app details: %w", err)
	}

	if len(apps) == 0 {
		return []*AdminAppDetails{}, nil
	}

	appIDs := make([]string, 0, len(apps))
	for _, a := range apps {
		appIDs = append(appIDs, a.ID)
	}

	componentsByApp, err := s.fetchComponentsWithLatestBuildByApp(ctx, appIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*AdminAppDetails, 0, len(apps))
	for _, a := range apps {
		items = append(items, &AdminAppDetails{
			App:        a,
			Components: componentsByApp[a.ID],
		})
	}

	return items, nil
}

func (s *service) fetchComponentsWithLatestBuildByApp(ctx context.Context, appIDs []string) (map[string][]AdminAppComponentDetails, error) {
	out := make(map[string][]AdminAppComponentDetails)
	if len(appIDs) == 0 {
		return out, nil
	}

	var components []*app.Component
	if err := s.db.WithContext(ctx).
		Where("app_id IN ?", appIDs).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("component_config_connections_latest_configs_view"))
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_builds.created_at DESC").Limit(1)
		}).
		Order("created_at asc").
		Find(&components).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch components for apps: %w", err)
	}

	for _, c := range components {
		out[c.AppID] = append(out[c.AppID], AdminAppComponentDetails{Component: c})
	}

	return out, nil
}
