package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// createShutdownHealthCheck writes a red health check to ClickHouse so dashboards
// immediately reflect the pending-shutdown state.
func (s *service) createShutdownHealthCheck(ctx context.Context, runnerID, processID string) {
	hc := app.RunnerHealthCheck{
		RunnerID:     runnerID,
		ProcessID:    processID,
		RunnerStatus: app.RunnerStatusError,
	}
	if res := s.chDB.WithContext(ctx).Create(&hc); res.Error != nil {
		s.l.Warn("unable to create shutdown health check", zap.Error(res.Error))
	}
}
