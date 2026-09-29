package settings

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nuonco/nuon/pkg/runner/version"
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
	s.LongPollJobs = settings.LongPollJobs
	s.TelemetryRelayEndpoint = settings.TelemetryRelayEndpoint
	s.VendorTelemetryEnabled = settings.VendorTelemetryEnabled
	s.VendorTelemetryResourceAttributes = settings.VendorTelemetryResourceAttributes
	s.EnableMetrics = settings.EnableMetrics
	s.EnableSentry = settings.EnableSentry
	s.Metadata = settings.Metadata
	s.EnableLogging = settings.EnableLogging
	s.LoggingLevel = level
	s.Groups = settings.Groups

	s.ContainerImageTag = settings.ContainerImageTag
	s.ContainerImageURL = settings.ContainerImageURL

	s.Metadata["runner.id"] = s.Cfg.RunnerID
	s.Metadata["runner.version"] = version.Version
	s.OtelSchemaURL = s.Cfg.RunnerAPIURL

	switch {
	case s.Cfg.RunnerPlatform != "":
		s.Platform = s.Cfg.RunnerPlatform
	case os.Getenv("CLOUD_PROVIDER") != "":
		s.Platform = os.Getenv("CLOUD_PROVIDER")
	case settings.AwsCloudformationStackType != "" || settings.AwsInstanceType != "":
		s.Platform = "aws"
	default:
		s.Platform = "aws"
	}

	return nil
}
