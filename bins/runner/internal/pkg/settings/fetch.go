package settings

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/powertoolsdev/mono/bins/runner/internal/version"
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
	s.Groups = settings.Groups

	// container
	s.ContainerImageTag = settings.ContainerImageTag
	s.ContainerImageURL = settings.ContainerImageURL

	// NOTE: we add a few additional fields into the metadata so they appear on all tags, but can not be set by the
	// API.
	s.Metadata["runner.id"] = s.Cfg.RunnerID
	s.Metadata["runner.version"] = version.Version
	s.OtelSchemaURL = s.Cfg.RunnerAPIURL

	// TODO(fd): return the platform on the RunnerGroupSettings object
	// platform
	s.Platform = "aws"
	if settings.AwsCloudformationStackType != "" && settings.AwsInstanceType != "" {
		s.Platform = "aws"
	}

	return nil
}
