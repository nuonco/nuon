package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetProcessLatestHeartBeat
// @Summary				get the latest heartbeat for a runner process
// @Description.markdown	get_runner_latest_heart_beat.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerHeartBeat
// @Router					/v1/runners/{runner_id}/processes/{process_id}/heart-beats/latest [get]
func (s *service) GetProcessLatestHeartBeat(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerID := ctx.Param("runner_id")
	processID := ctx.Param("process_id")

	_, err = s.getOrgRunner(ctx, runnerID, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	heartBeat, err := s.getProcessLatestHeartBeat(ctx, runnerID, processID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, heartBeat)
}

func (s *service) getProcessLatestHeartBeat(ctx *gin.Context, runnerID, processID string) (*app.RunnerHeartBeat, error) {
	var heartBeat app.RunnerHeartBeat

	resp := s.chDB.WithContext(ctx).
		Where("runner_id = ? AND process_id = ?", runnerID, processID).
		Order("created_at desc").
		Limit(1).
		First(&heartBeat)

	if resp.Error != nil {
		return nil, resp.Error
	}

	return &heartBeat, nil
}
