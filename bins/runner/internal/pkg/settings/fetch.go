package settings

import (
	"context"
	"fmt"
	"time"
)

func (s *Settings) fetch(ctx context.Context) error {
	settings, err := s.apiClient.GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("unable to get settings: %w", err)
	}

	s.HeartBeatTimeout = time.Duration(settings.HeartBeatTimeout)

	return nil
}
