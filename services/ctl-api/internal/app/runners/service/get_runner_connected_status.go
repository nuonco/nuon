package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	heartBeatConnectCheckWindowSeconds = 15
)

type RunnerConnectionStatus struct {
	Connected                bool  `json:"connected"`
	LatestHeartbeatTimestamp int64 `json:"latest_heartbeat_timestamp"`
}

// @ID						GetRunnerConnectStatus
// @Summary					get a runner connection satus based on heartbeat
// @Description.markdown	get_runner_connect_status.md
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	RunnerConnectionStatus
// @Router					/v1/runners/{runner_id}/connected [get]
func (s *service) GetRunnerConnectStatus(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerID := ctx.Param("runner_id")

	_, err = s.getOrgRunner(ctx, runnerID, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	now := time.Now()
	hb, err := s.getRunnerLatestHeartBeat(ctx, runnerID)
	if err != nil {
		ctx.JSON(http.StatusOK, RunnerConnectionStatus{
			Connected: false,
		})
		return
	}

	if now.Unix()-hb.CreatedAt.Unix() > heartBeatConnectCheckWindowSeconds {
		ctx.JSON(http.StatusOK, RunnerConnectionStatus{
			Connected:                false,
			LatestHeartbeatTimestamp: hb.CreatedAt.Unix(),
		})
		return
	}

	ctx.JSON(http.StatusOK, RunnerConnectionStatus{
		Connected:                true,
		LatestHeartbeatTimestamp: hb.CreatedAt.Unix(),
	})
}
