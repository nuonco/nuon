package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"
)

func (s *sync) syncApp(ctx context.Context, resource string) error {
	_, err := s.apiClient.UpdateApp(ctx, s.appID, &models.ServiceUpdateAppRequest{
		Description:     s.cfg.Description,
		DisplayName:     s.cfg.DisplayName,
		SlackWebhookURL: s.cfg.SlackWebhookURL,
	})
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
