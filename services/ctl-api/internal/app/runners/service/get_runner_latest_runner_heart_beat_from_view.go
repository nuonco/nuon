package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type LatestRunnerHeartBeats map[string]*app.LatestRunnerHeartBeat

// @ID						GetLatestRunnerHeartBeat
// @Summary				get a runner
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
// @Success				200	{object}	LatestRunnerHeartBeats
// @Router					/v1/runners/{runner_id}/heart-beats/latest [get]
func (s *service) GetLatestRunnerHeartBeatFromView(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	heartBeats, err := s.getRunnerLatestHeartBeatFromView(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, heartBeats)
}

func (s *service) getRunnerLatestHeartBeatFromView(ctx context.Context, runnerID string) (LatestRunnerHeartBeats, error) {
	var runnerHeartBeats []*app.LatestRunnerHeartBeat

	resp := s.chDB.WithContext(ctx).
		Where("runner_id = ?", runnerID).
		Order("created_at_latest desc").
		Find(&runnerHeartBeats)

	if resp.Error != nil {
		return nil, resp.Error
	}

	// NOTE(fd): the view de-dupes but that's eventually consistent so we're going to dedupe here as we compose the repsonse.
	heartbeats := LatestRunnerHeartBeats{}
	for _, rhb := range runnerHeartBeats {
		process := string(rhb.Process)
		value, ok := heartbeats[process]
		if !ok { // not found: add
			heartbeats[process] = rhb
		} else if rhb.CreatedAt.After(value.CreatedAt) {
			heartbeats[process] = rhb
		}
	}

	return heartbeats, nil
}
