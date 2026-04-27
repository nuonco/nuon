package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type LatestRunnerProcessHeartBeats map[string]*app.LatestRunnerHeartBeatV2

// @ID						GetLatestRunnerProcessHeartBeat
// @Summary				get a runner
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
// @Success				200	{object}	LatestRunnerProcessHeartBeats
// @Router					/v1/runners/{runner_id}/processes/{process_id}/heart-beats/latest [get]
func (s *service) GetLatestRunnerProcessHeartBeatFromView(ctx *gin.Context) {
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

	heartBeats, err := s.getRunnerProcessLatestHeartBeatFromView(ctx, runnerID, processID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, heartBeats)
}

func (s *service) getRunnerProcessLatestHeartBeatFromView(ctx context.Context, runnerID, processID string) (LatestRunnerProcessHeartBeats, error) {
	var runnerHeartBeats []*app.LatestRunnerHeartBeatV2

	resp := s.chDB.WithContext(ctx).
		Where("runner_id = ? AND process_id = ?", runnerID, processID).
		Order("created_at_latest desc").
		Find(&runnerHeartBeats)

	if resp.Error != nil {
		return nil, resp.Error
	}

	// NOTE(fd): the view de-dupes but that's eventually consistent so we're going to dedupe here as we compose the response.
	heartbeats := LatestRunnerProcessHeartBeats{}
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
