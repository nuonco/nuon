package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type AdminInstallDetails struct {
	*app.Install

	Components []AdminInstallComponentDetails `json:"components,omitempty"`
}

type AdminInstallComponentDetails struct {
	*app.InstallComponent

	LatestDeploy *app.InstallDeploy `json:"latest_deploy,omitempty"`
}

const adminInstallDetailsDefaultLimit = 25

// @ID			AdminListInstallsDetails
// @BasePath	/v1/installs
// @Summary	Return a compact admin list of installs with their components and latest deploy status
// @Description	Admin list of installs intended for status / README rollups.
// @Description	Each install includes its components and each component's most
// @Description	recent deploy, whose `status_description` surfaces actionable
// @Description	error messages on failure.
// @Description	Pagination is uncapped on this admin endpoint — pass any `limit`.
// @Description	The optional `status` query parameter filters installs that have at
// @Description	least one component whose `status_v2->>'status'` matches. The
// @Description	parameter may be repeated (e.g. `?status=error&status=pending`).
// @Param			offset	query	int			false	"offset of results to return"	Default(0)
// @Param			limit	query	int			false	"limit of results to return (no upper cap)"	Default(25)
// @Param			status	query	[]string	false	"filter installs by component composite status (repeatable)"	collectionFormat(multi)
// @Tags			installs/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array}	AdminInstallDetails
// @Router			/v1/installs/details [GET]
func (s *service) AdminListInstallsDetails(ctx *gin.Context) {
	limit, offset := parseAdminInstallDetailsPagination(ctx)
	statuses := ctx.QueryArray("status")

	items, err := s.listInstallsDetails(ctx, limit, offset, statuses)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, items)
}

func parseAdminInstallDetailsPagination(ctx *gin.Context) (limit, offset int) {
	limit = adminInstallDetailsDefaultLimit
	if v, err := strconv.Atoi(ctx.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(ctx.Query("offset")); err == nil && v >= 0 {
		offset = v
	}
	return limit, offset
}

func (s *service) listInstallsDetails(ctx *gin.Context, limit, offset int, statuses []string) ([]*AdminInstallDetails, error) {
	var installs []*app.Install
	tx := s.db.WithContext(ctx).
		Preload("AppRunnerConfig").
		Order(views.TableOrViewName(s.db, &app.Install{}, ".created_at DESC")).
		Limit(limit).
		Offset(offset)
	if len(statuses) > 0 {
		installView := views.TableOrViewName(s.db, &app.Install{}, "")
		tx = tx.Where(
			"EXISTS (SELECT 1 FROM install_components ic WHERE ic.install_id = "+installView+
				".id AND ic.deleted_at = 0 AND ic.status_v2->>'status' IN ?)",
			statuses,
		)
	}
	if err := tx.Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to list install details: %w", err)
	}

	if len(installs) == 0 {
		return []*AdminInstallDetails{}, nil
	}

	installIDs := make([]string, 0, len(installs))
	for _, i := range installs {
		installIDs = append(installIDs, i.ID)
	}

	componentsByInstall, err := s.fetchInstallComponentsWithLatestDeploy(ctx, installIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*AdminInstallDetails, 0, len(installs))
	for _, i := range installs {
		items = append(items, &AdminInstallDetails{
			Install:    i,
			Components: componentsByInstall[i.ID],
		})
	}

	return items, nil
}

func (s *service) fetchInstallComponentsWithLatestDeploy(ctx context.Context, installIDs []string) (map[string][]AdminInstallComponentDetails, error) {
	out := make(map[string][]AdminInstallComponentDetails)
	if len(installIDs) == 0 {
		return out, nil
	}

	var installComponents []*app.InstallComponent
	if err := s.db.WithContext(ctx).
		Where("install_id IN ?", installIDs).
		Preload("Component").
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC").Limit(1)
		}).
		Order("created_at asc").
		Find(&installComponents).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch install components: %w", err)
	}

	for _, ic := range installComponents {
		summary := AdminInstallComponentDetails{InstallComponent: ic}
		if len(ic.InstallDeploys) > 0 {
			summary.LatestDeploy = &ic.InstallDeploys[0]
		}
		out[ic.InstallID] = append(out[ic.InstallID], summary)
	}

	return out, nil
}
