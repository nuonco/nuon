package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SandboxRunnerInfo struct {
	Runner                   app.Runner `json:"runner"`
	Connected                bool       `json:"connected"`
	LatestHeartbeatTimestamp int64      `json:"latest_heartbeat_timestamp"`
}

func (s *service) AdminListSandboxRunners(ctx *gin.Context) {
	runners, err := s.listSandboxRunners(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list sandbox runners: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runners)
}

func (s *service) listSandboxRunners(ctx context.Context) ([]SandboxRunnerInfo, error) {
	// Find all runner groups with sandbox mode enabled
	var settings []app.RunnerGroupSettings
	if res := s.db.WithContext(ctx).
		Where(app.RunnerGroupSettings{SandboxMode: true}).
		Find(&settings); res.Error != nil {
		return nil, fmt.Errorf("unable to find sandbox settings: %w", res.Error)
	}

	if len(settings) == 0 {
		return []SandboxRunnerInfo{}, nil
	}

	// Collect runner group IDs
	groupIDs := make([]string, len(settings))
	for i, s := range settings {
		groupIDs[i] = s.RunnerGroupID
	}

	// Find all runners in those groups
	var runners []app.Runner
	if res := s.db.WithContext(ctx).
		Where("runner_group_id IN ?", groupIDs).
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to find sandbox runners: %w", res.Error)
	}

	// Check heartbeat for each runner
	now := time.Now()
	result := make([]SandboxRunnerInfo, 0, len(runners))
	for _, runner := range runners {
		info := SandboxRunnerInfo{
			Runner: runner,
		}

		hb, err := s.getRunnerLatestHeartBeat(ctx, runner.ID)
		if err == nil {
			info.LatestHeartbeatTimestamp = hb.CreatedAt.Unix()
			info.Connected = now.Unix()-hb.CreatedAt.Unix() < heartBeatConnectCheckWindowSeconds
		}

		result = append(result, info)
	}

	return result, nil
}
