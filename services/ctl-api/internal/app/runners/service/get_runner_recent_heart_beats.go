package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// Result represents a rolled up heart every 2 minutes
type RunnerRecentHeartBeatResponse struct {
	RunnerID      string    `json:"runner_id"`
	TruncatedTime time.Time `json:"truncated_time"`
	RecordCount   int       `json:"record_count"`
}

// @ID GetRunnerRecentHeartBeats
// @Summary	get a runner
// @Description.markdown get_runner_recent_heart_beats.md
// @Param			runner_id	path	string	true	"runner ID"
// @Tags runners
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}	RunnerRecentHeartBeatResponse
// @Router			/v1/runners/{runner_id}/recent-heart-beats [get]
func (s *service) GetRunnerRecentHeartBeats(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	_, err := s.getRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	heartBeats, err := s.getRunnerRecentHeartBeats(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, heartBeats)
}

func (s *service) getRunnerRecentHeartBeats(ctx context.Context, runnerID string) ([]*RunnerRecentHeartBeatResponse, error) {
	runnerHeartBeat := app.RunnerHeartBeat{}

	var results []*RunnerRecentHeartBeatResponse

	err := s.chDB.WithContext(ctx).
		Model(&runnerHeartBeat).
		Select("runner_id, DATE_TRUNC('minute', created_at) AS truncated_time, COUNT(*) AS record_count").
		Where("runner_id = ? AND EXTRACT(MINUTE FROM created_at) % 2 = 0 AND created_at > NOW() - INTERVAL '1 hour'", runnerID).
		Group("runner_id, truncated_time").
		Order("truncated_time ASC").
		Scan(&results).Error

	if results == nil {
		return []*RunnerRecentHeartBeatResponse{}, err
	}

	return results, err
}
