package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) ValidateCloudConnection(ctx context.Context, orgID, connectionID string, platform app.CloudPlatform) (*app.CloudConnection, error) {
	var connection app.CloudConnection
	result := s.db.WithContext(ctx).Where(&app.CloudConnection{ID: connectionID, OrgID: orgID}).First(&connection)
	if result.Error != nil {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection not found: %w", result.Error), Description: "Cloud connection was not found for this organization"}
	}
	if connection.Platform != app.CloudPlatformAWS || connection.Platform != platform {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection %s is for %s, not %s", connection.ID, connection.Platform, platform), Description: "Cloud connection does not match the app platform"}
	}
	if connection.Status != app.CloudConnectionStatusVerified {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection %s is not verified", connection.ID), Description: "Cloud connection must be verified before it can be used"}
	}
	if connection.Preset != app.CloudConnectionPresetStacks && connection.Preset != app.CloudConnectionPresetCustom {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection %s has an unsupported preset", connection.ID), Description: "Cloud connection preset must be stacks or custom"}
	}
	return &connection, nil
}
