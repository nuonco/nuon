package settings

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (s *Settings) fetch(ctx context.Context) error {
	settings, err := s.apiClient.GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("unable to get settings: %w", err)
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(settings.LoggingLevel)); err != nil {
		return fmt.Errorf("unable to parse logging level: %w", err)
	}

	s.HeartBeatTimeout = time.Duration(settings.HeartBeatTimeout)
	s.SandboxMode = settings.SandboxMode
	s.EnableMetrics = settings.EnableMetrics
	s.EnableSentry = settings.EnableSentry
	s.Metadata = settings.Metadata
	s.EnableLogging = settings.EnableLogging
	s.LoggingLevel = level

	return nil
}
