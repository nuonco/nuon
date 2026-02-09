package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetOrgRunnerGroup
// @Summary				Get an org's runner group
// @Description.markdown	get_org_runner_group.md
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
// @Success				200	{object}	app.RunnerGroup
// @Router					/v1/orgs/current/runner-group [GET]
func (s *service) GetOrgRunnerGroup(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runnerGroup, err := s.getOrgRunnerGroup(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runnerGroup)
}

func (s *service) getOrgRunnerGroup(ctx context.Context, orgID string) (*app.RunnerGroup, error) {
	runnerGroup := app.RunnerGroup{}
	res := s.db.WithContext(ctx).
		Preload("Runners").
		Preload("Settings").
		Where(app.RunnerGroup{
			OwnerType: "orgs",
			OwnerID:   orgID,
		}).
		First(&runnerGroup)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("runner group not found"),
				Description: fmt.Sprintf("runner group for org %s not found", orgID),
			}
		}
		return nil, fmt.Errorf("unable to get org runner group %s: %w", orgID, res.Error)
	}

	return &runnerGroup, nil
}
