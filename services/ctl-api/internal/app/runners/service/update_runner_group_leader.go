package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runnergroupssignals "github.com/nuonco/nuon/services/ctl-api/internal/app/runner_groups/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type updateRunnerGroupLeaderRequest struct {
	RunnerID *string `json:"runner_id"`
}

// @ID						UpdateRunnerGroupLeader
// @Summary				set or auto-elect the leader runner for a runner group
// @Tags					runners
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
		// Trigger async leader election via the runner-groups event loop.
		s.evClient.Send(ctx, groupID, &runnergroupssignals.Signal{
			Type: runnergroupssignals.OperationElectLeader,
		})
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

		// Clear all leader flags in the group, then set the requested runner as leader.
		if err := s.db.WithContext(ctx).Model(&app.Runner{}).
			Where("runner_group_id = ? AND org_id = ? AND deleted_at = 0", groupID, org.ID).
			Update("leader", false).Error; err != nil {
			ctx.Error(fmt.Errorf("unable to clear leader flags: %w", err))
			return
		}
		if err := s.db.WithContext(ctx).Model(&app.Runner{}).
			Where("id = ? AND deleted_at = 0", *req.RunnerID).
			Update("leader", true).Error; err != nil {
			ctx.Error(fmt.Errorf("unable to set leader: %w", err))
			return
		}
	}

	// Find and return the current leader runner.
	var leader app.Runner
	res := s.db.WithContext(ctx).
		Where("runner_group_id = ? AND leader = true AND deleted_at = 0", groupID).
		First(&leader)
	if res.Error != nil {
		ctx.JSON(http.StatusOK, gin.H{"leader_runner_id": nil})
		return
	}

	ctx.JSON(http.StatusOK, &leader)
}
