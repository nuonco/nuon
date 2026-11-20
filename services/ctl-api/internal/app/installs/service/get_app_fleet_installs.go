package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
)

const fleetDefaultPageLimit = 50

// @ID						GetAppFleetInstalls
// @Summary				paginated installs for an app fleet
// @Description			Returns the app's installs matching the fleet filters, ordered by rollout, health severity, then name. Pass app_branch_id and app_branch_group to page through one install group, for example when a lane card's panel is opened.
// @Param					app_id			path	string	true	"app ID"
// @Param					q				query	string	false	"search by install name or ID"
// @Param					labels			query	string	false	"label filter (key:value,key:value; value * matches key presence)"
// @Param					health			query	string	false	"comma-separated health: unhealthy,degraded,unknown,progressing,healthy,no-signal"
// @Param					runner_status	query	string	false	"runner filter: active or not-active"
// @Param					rollout			query	string	false	"rollout filter: applied, behind, or unknown"
// @Param					component_id	query	string	false	"compare rollout on one component's applied config instead of the whole install"
// @Param					app_branch_id	query	string	false	"limit to one branch; __none__ for installs without a branch"
// @Param					app_branch_group	query	string	false	"limit to one install group; __none__ for installs without a group"
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(50)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
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
// @Success				200	{array}		service.AppFleetInstall
// @Router					/v1/apps/{app_id}/fleet/installs [GET]
func (s *service) GetAppFleetInstalls(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	currentApp, err := s.findAppByNameOrID(ctx, org.ID, ctx.Param("app_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	filters, err := parseFleetFilters(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	rows, err := s.getAppFleetInstalls(ctx, org.ID, currentApp.ID, filters)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app fleet installs: %w", err))
		return
	}

	out := make([]AppFleetInstall, 0, len(rows))
	for _, row := range rows {
		out = append(out, fleetInstallFromRow(row))
	}

	ctx.JSON(http.StatusOK, out)
}

func (s *service) getAppFleetInstalls(ctx *gin.Context, orgID, appID string, f fleetFilters) ([]fleetInstallRow, error) {
	limit, offset := fleetDefaultPageLimit, 0
	if page := cctx.OffsetPaginationFromContext(ctx); page != nil {
		limit, offset = page.Limit, page.Offset
	}

	var rows []fleetInstallRow
	err := s.fleetTx(ctx, func(tx *gorm.DB) error {
		branches, err := loadFleetBranches(tx, orgID, appID)
		if err != nil {
			return err
		}
		targetsJSON, err := fleetTargetsJSON(branches.targets)
		if err != nil {
			return err
		}

		sql, args, err := fleetInstallsSQL(fleetViewName(tx), targetsJSON, orgID, appID, f)
		if err != nil {
			return err
		}
		args = append(args, limit+1, offset)

		if err := tx.Raw(sql, args...).Scan(&rows).Error; err != nil {
			return fmt.Errorf("unable to list fleet installs: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return db.HandlePaginatedResponse(ctx, rows)
}
