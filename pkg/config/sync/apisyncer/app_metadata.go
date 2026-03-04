package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) syncApp(ctx context.Context, resource string) error {
	_, err := s.apiClient.UpdateApp(ctx, s.appID, &models.ServiceUpdateAppRequest{
		// NOTE: we don't allow changing the ServiceUpdateAppRequest.app name from the config
		Description:     s.cfg.Description,
		DisplayName:     s.cfg.DisplayName,
		SlackWebhookURL: s.cfg.SlackWebhookURL,
	})
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
