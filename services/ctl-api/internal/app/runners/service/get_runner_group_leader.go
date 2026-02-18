package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetRunnerGroupLeader
// @Summary				get the leader runner for a runner group
// @Tags					runners
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Param					runner_group_id	path	string	true	"runner group ID"
// @Success				200	{object}	app.Runner
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/runner-groups/{runner_group_id}/leader [get]
func (s *service) GetRunnerGroupLeader(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	groupID := ctx.Param("runner_group_id")

	var leader app.Runner
	res := s.db.WithContext(ctx).
		Where("runner_group_id = ? AND org_id = ? AND leader = true AND deleted_at = 0", groupID, org.ID).
		First(&leader)
	if res.Error != nil {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("no leader elected for runner group %s", groupID),
			Description: "no leader elected",
		})
		return
	}

	ctx.JSON(http.StatusOK, &leader)
}
