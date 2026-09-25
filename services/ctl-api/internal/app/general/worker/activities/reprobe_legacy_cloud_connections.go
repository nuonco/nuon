package activities

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/service"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
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
	if err := a.db.WithContext(ctx).Find(&connections).Error; err != nil {
		return nil, err
	}
	response := &ReprobeLegacyCloudConnectionsResponse{}
	for i := range connections {
		connection := &connections[i]
		response.Probed++
		previousStatus := connection.Status
		result, verifyErr := a.cloudConnectionVerifier.Verify(ctx, connection, service.VerifyOptions{})
		now := time.Now().UTC()
		if verifyErr != nil {
			response.Failures++
			result.Status = app.CloudConnectionStatusError
			result.Message = fmt.Sprintf("Cloud connection verification failed: %v", verifyErr)
		}
		updates := map[string]any{
			"status":           result.Status,
			"status_message":   result.Message,
			"last_verified_at": &now,
			"capabilities":     result.Capabilities,
			"registries":       result.Registries,
		}
		if result.Status == app.CloudConnectionStatusVerified {
			updates["auth_mode"] = app.CloudConnectionAuthModeOIDC
		}
		if err := a.db.WithContext(ctx).Model(connection).Updates(updates).Error; err != nil {
			response.Failures++
			a.l.Warn("cloud connection re-verification update failed", zap.String("connection_id", connection.ID), zap.Error(err))
			continue
		}
		if previousStatus == app.CloudConnectionStatusVerified && result.Status != app.CloudConnectionStatusVerified {
			err := a.orgsHelpers.EnqueueOrgSignal(ctx, orgshelpers.EnqueueOrgSignalParams{
				OrgID: connection.OrgID,
				Signal: &verificationfailed.Signal{
					ConnectionID: connection.ID, ConnectionName: connection.Name, OrgID: connection.OrgID,
					Platform: connection.Platform, Message: result.Message,
				},
				IdempotencyKey: fmt.Sprintf("cloud-connection-verification-failed-%s-%s", connection.ID, now.Format(time.RFC3339Nano)),
			})
			if err != nil {
				response.Failures++
				a.l.Warn("cloud connection verification failure signal enqueue failed", zap.String("connection_id", connection.ID), zap.Error(err))
			}
		}
		if result.Status == app.CloudConnectionStatusVerified {
			response.OIDC++
		} else if connection.AuthMode == app.CloudConnectionAuthModeLegacy || connection.AuthMode == "" {
			response.Legacy++
			a.l.Info("cloud connection still uses legacy authentication", zap.String("connection_id", connection.ID), zap.String("org_id", connection.OrgID), zap.String("platform", string(connection.Platform)))
		}
	}
	a.mw.Count("cloud_connections.legacy_reprobe", int64(response.Probed), []string{})
	a.mw.Count("cloud_connections.legacy_remaining", int64(response.Legacy), []string{})
	return response, nil
}
