package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) ValidateCloudConnection(ctx context.Context, orgID, connectionID string, platform app.CloudPlatform, capability app.CloudConnectionCapability) (*app.CloudConnection, error) {
	var connection app.CloudConnection
	result := s.db.WithContext(ctx).Where(&app.CloudConnection{ID: connectionID, OrgID: orgID}).First(&connection)
	if result.Error != nil {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection not found: %w", result.Error), Description: "Cloud connection was not found for this organization"}
	}
	if connection.Platform != platform {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection %s is for %s, not %s", connection.ID, connection.Platform, platform), Description: "Cloud connection does not match the app platform"}
	}
	if connection.Status != app.CloudConnectionStatusVerified {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection %s is not verified", connection.ID), Description: "Cloud connection must be verified before it can be used"}
	}
	if !connection.HasCapability(capability) {
		return nil, stderr.ErrUser{Err: fmt.Errorf("cloud connection %s lacks %s capability", connection.ID, capability), Description: fmt.Sprintf("Cloud connection must have the %s capability", capability)}
	}
	return &connection, nil
}
