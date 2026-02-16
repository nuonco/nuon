package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type updateRunnerGroupLeaderRequest struct {
	RunnerID *string `json:"runner_id"`
}

// @ID						UpdateRunnerGroupLeader
// @Summary				set or auto-elect the leader runner for a runner group
// @Tags					runners/runner-groups
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Param					runner_group_id	path	string								true	"runner group ID"
// @Param					request			body	updateRunnerGroupLeaderRequest		true	"leader update request"
// @Success				200	{object}	app.Runner
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/runner-groups/{runner_group_id}/leader [put]
func (s *service) UpdateRunnerGroupLeader(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	groupID := ctx.Param("runner_group_id")

	var req updateRunnerGroupLeaderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid request: %w", err),
			Description: "invalid request body",
		})
		return
	}

	if req.RunnerID == nil {
		if err := s.helpers.ElectLeader(ctx, groupID); err != nil {
			ctx.Error(fmt.Errorf("unable to auto-elect leader: %w", err))
			return
		}
	} else {
		var runner app.Runner
		res := s.db.WithContext(ctx).First(&runner, "id = ? AND runner_group_id = ? AND org_id = ?", *req.RunnerID, groupID, org.ID)
		if res.Error != nil {
			ctx.Error(stderr.ErrNotFound{
				Err:         fmt.Errorf("runner %s not found in group %s: %w", *req.RunnerID, groupID, res.Error),
				Description: "runner not found in this group",
			})
			return
		}

		if runner.Status != app.RunnerStatusActive {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("runner %s is not active (status: %s)", *req.RunnerID, runner.Status),
				Description: "runner must be active to be elected leader",
			})
			return
		}

		res = s.db.WithContext(ctx).Model(&app.RunnerGroup{}).Where("id = ? AND org_id = ?", groupID, org.ID).Update("leader_runner_id", *req.RunnerID)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to set leader: %w", res.Error))
			return
		}
	}

	rg := app.RunnerGroup{}
	res := s.db.WithContext(ctx).
		Preload("LeaderRunner").
		First(&rg, "id = ? AND org_id = ?", groupID, org.ID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runner group: %w", res.Error))
		return
	}

	if rg.LeaderRunner == nil {
		ctx.JSON(http.StatusOK, gin.H{"leader_runner_id": nil})
		return
	}

	ctx.JSON(http.StatusOK, rg.LeaderRunner)
}
