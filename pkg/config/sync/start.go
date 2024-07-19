package sync

import (
	"context"
)

func (s *sync) start(ctx context.Context) error {
	cfg, err := s.apiClient.CreateAppConfig(ctx, s.appID, nil)
	if err != nil {
		return err
	}

	s.state.CfgID = cfg.ID
	return nil
}
