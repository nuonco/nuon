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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type ReverifyCloudConnectionsResponse struct {
	Probed   int `json:"probed"`
	OIDC     int `json:"oidc"`
	Legacy   int `json:"legacy"`
	Failures int `json:"failures"`
}

type ReverifyCloudConnectionsRequest struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30m
func (a *Activities) ReverifyCloudConnections(ctx context.Context, _ ReverifyCloudConnectionsRequest) (*ReverifyCloudConnectionsResponse, error) {
	var connections []app.CloudConnection
	if err := a.db.WithContext(ctx).Find(&connections).Error; err != nil {
		return nil, err
	}
	response := &ReverifyCloudConnectionsResponse{}
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
			a.l.Warn("cloud connection re-verification failed", zap.String("connection_id", connection.ID), zap.Error(verifyErr))
		}
		update := app.CloudConnection{Status: result.Status, StatusMessage: result.Message, LastVerifiedAt: &now, Capabilities: result.Capabilities, Registries: connection.Registries}
		selected := []string{"status", "status_message", "last_verified_at", "capabilities", "registries"}
		if result.Status == app.CloudConnectionStatusVerified {
			update.AuthMode = app.CloudConnectionAuthModeOIDC
			update.Registries = result.Registries
			selected = append(selected, "auth_mode")
		}
		if err := a.db.WithContext(ctx).Model(&app.CloudConnection{}).Where(app.CloudConnection{OrgID: connection.OrgID, ID: connection.ID, Principal: connection.Principal}).Select(selected).Updates(update).Error; err != nil {
			response.Failures++
			a.l.Warn("cloud connection re-verification update failed", zap.String("connection_id", connection.ID), zap.Error(err))
			continue
		}
		if previousStatus == app.CloudConnectionStatusVerified && result.Status != app.CloudConnectionStatusVerified {
			// queue_signals.created_by_id is NOT NULL and filled from context. An
			// activity has no account, so attribute the signal to whoever created
			// the connection, the way the component health sweep does.
			signalCtx := context.WithValue(ctx, keys.AccountIDCtxKey, connection.CreatedByID)
			err := a.orgsHelpers.EnqueueOrgSignal(signalCtx, orgshelpers.EnqueueOrgSignalParams{
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
	a.mw.Count("cloud_connections.reverified", int64(response.Probed), []string{})
	a.mw.Count("cloud_connections.legacy_remaining", int64(response.Legacy), []string{})
	return response, nil
}
