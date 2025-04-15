package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s *sync) start(ctx context.Context) error {
	cfg, err := s.apiClient.CreateAppConfig(ctx, s.appID, &models.ServiceCreateAppConfigRequest{
		Readme: s.cfg.Readme,
	})
	if err != nil {
		return err
	}

	s.appConfigID = cfg.ID
	s.state.CfgID = cfg.ID
	return nil
}
