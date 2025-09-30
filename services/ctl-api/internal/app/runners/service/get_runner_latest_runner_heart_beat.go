package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetRunnerLatestHeartBeat
// @Summary				get the latest heartbeats for a runner
// @Description.markdown	get_runner_latest_heart_beat.md
// @Param					runner_id	path	string	true	"runner ID"
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
// @Router					/v1/runners/{runner_id}/latest-heart-beat [get]
func (s *service) GetRunnerLatestHeartBeat(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	heartBeats, err := s.getRunnerLatestHeartBeat(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, heartBeats)
}

func (s *service) getRunnerLatestHeartBeat(ctx context.Context, runnerID string) (*app.RunnerHeartBeat, error) {
	var runnerHeartBeat app.RunnerHeartBeat

	resp := s.chDB.WithContext(ctx).
		Where("runner_id = ?", runnerID).
		Order("created_at desc").
		Limit(1).
		First(&runnerHeartBeat)

	if resp.Error != nil {
		return nil, resp.Error
	}

	return &runnerHeartBeat, nil
}
