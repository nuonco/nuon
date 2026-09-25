package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetCloudConnectionCredentialsRequest struct {
	ConnectionID string
	SessionName  string
}

// @temporal-gen-v2 activity
// @max-retries 1
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 45s
func (a *Activities) GetCloudConnectionCredentials(ctx context.Context, req *GetCloudConnectionCredentialsRequest) (*credentials.Config, error) {
	var connection app.CloudConnection
	if err := a.db.WithContext(ctx).Where(app.CloudConnection{ID: req.ConnectionID}).First(&connection).Error; err != nil {
		return nil, fmt.Errorf("get cloud connection: %w", err)
	}
	return a.cloudConnections.ECRCredentials(ctx, &connection, req.SessionName)
}
