package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetCloudConnectionGARAuthRequest struct {
	ConnectionID string
}

// @temporal-gen-v2 activity
// @max-retries 1
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 45s
func (a *Activities) GetCloudConnectionGARAuth(ctx context.Context, req *GetCloudConnectionGARAuthRequest) (*configs.OCIRegistryAuth, error) {
	var connection app.CloudConnection
	if err := a.db.WithContext(ctx).Where(app.CloudConnection{ID: req.ConnectionID}).First(&connection).Error; err != nil {
		return nil, fmt.Errorf("get cloud connection: %w", err)
	}
	token, err := a.cloudConnections.GCPAccessToken(ctx, &connection)
	if err != nil {
		return nil, err
	}
	return &configs.OCIRegistryAuth{Username: "oauth2accesstoken", Password: token.AccessToken}, nil
}
