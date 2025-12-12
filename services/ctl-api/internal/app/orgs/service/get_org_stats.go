package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type OrgStatsResponse struct {
	InstallNames []string `json:"install_names"`
	AppCount     int64    `json:"app_count"`
	InstallCount int64    `json:"install_count"`
}

// @ID                     GetOrgStats
// @Summary				Get an org
// @Description.markdown	get_org_stats.md
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/current/stats [GET]
func (s *service) GetOrgStats(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	stats, err := s.getOrgStats(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

func (s *service) getOrgStats(ctx context.Context, orgID string) (*OrgStatsResponse, error) {

	var installCount int64
	var appCount int64
	fmt.Println("Getting stats for org:", orgID)
	s.db.WithContext(ctx).
		Model(&app.App{}).
		Where("org_id = ?", orgID).
		Count(&appCount)

	s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("org_id = ?", orgID).
		Count(&installCount)

	fmt.Println("AppCount:", appCount, "InstallCount:", installCount)

	installNames := []string{}
	rows, err := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Select("installs.name").
		Where("org_id = ?", orgID).
		Rows()
	if err != nil {
		return nil, fmt.Errorf("unable to get installation names for org %s: %w", orgID, err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("unable to scan installation name for org %s: %w", orgID, err)
		}
		installNames = append(installNames, name)
	}

	return &OrgStatsResponse{
		InstallNames: installNames,
		AppCount:     appCount,
		InstallCount: installCount,
	}, nil
}
