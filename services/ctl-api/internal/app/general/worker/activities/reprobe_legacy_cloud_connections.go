package activities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReprobeLegacyCloudConnectionsResponse struct {
	Probed   int `json:"probed"`
	OIDC     int `json:"oidc"`
	Legacy   int `json:"legacy"`
	Failures int `json:"failures"`
}

type ReprobeLegacyCloudConnectionsRequest struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30m
func (a *Activities) ReprobeLegacyCloudConnections(ctx context.Context, _ ReprobeLegacyCloudConnectionsRequest) (*ReprobeLegacyCloudConnectionsResponse, error) {
	var connections []app.CloudConnection
	if err := a.db.WithContext(ctx).
		Where(app.CloudConnection{Platform: app.CloudPlatformAWS}).
		Where(map[string]any{"auth_mode": []app.CloudConnectionAuthMode{"", app.CloudConnectionAuthModeLegacy}}).
		Find(&connections).Error; err != nil {
		return nil, err
	}
	response := &ReprobeLegacyCloudConnectionsResponse{}
	for i := range connections {
		response.Probed++
		if _, err := a.cloudConnectionsHelpers.ECRCredentials(ctx, &connections[i], "nuon-legacy-reprobe"); err != nil {
			response.Failures++
			a.l.Warn("cloud connection legacy auth reprobe failed", zap.String("connection_id", connections[i].ID), zap.Error(err))
			continue
		}
		if connections[i].AuthMode == app.CloudConnectionAuthModeOIDC {
			response.OIDC++
		} else {
			response.Legacy++
			a.l.Info("cloud connection still uses legacy authentication", zap.String("connection_id", connections[i].ID), zap.String("org_id", connections[i].OrgID))
		}
	}
	a.mw.Count("cloud_connections.legacy_reprobe", int64(response.Probed), []string{})
	a.mw.Count("cloud_connections.legacy_remaining", int64(response.Legacy), []string{})
	return response, nil
}
