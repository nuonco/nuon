package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type ReverifyRequest struct {
	CloudConnectionID string `json:"cloud_connection_id" validate:"required"`
	OnDemand          bool   `json:"on_demand,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) Reverify(ctx context.Context, req ReverifyRequest) error {
	if !req.OnDemand {
		return nil
	}
	var connection app.CloudConnection
	if err := a.db.WithContext(ctx).Where(app.CloudConnection{ID: req.CloudConnectionID}).First(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.l.Debug("skipping deleted cloud connection", zap.String("connection_id", req.CloudConnectionID))
			return nil
		}
		return err
	}
	result, verifyErr := a.verifier.Verify(ctx, &connection, cloudconnections.VerifyOptions{RetryIAMPropagation: connection.LastVerifiedAt == nil})
	if verifyErr != nil {
		result = cloudconnections.VerificationResult{Status: app.CloudConnectionStatusError, Message: cloudconnections.VerificationErrorMessage(verifyErr)}
	}
	if verifyErr != nil {
		a.l.Warn("cloud connection re-verification failed", zap.String("connection_id", connection.ID), zap.Error(verifyErr))
	}
	now := time.Now().UTC()
	update := app.CloudConnection{Status: result.Status, StatusMessage: result.Message, LastVerifiedAt: &now}
	res := a.db.WithContext(ctx).Model(&app.CloudConnection{}).
		Where(app.CloudConnection{OrgID: connection.OrgID, ID: connection.ID, Principal: connection.Principal}).Select([]string{"status", "status_message", "last_verified_at"}).Updates(update)
	if res.Error != nil {
		return fmt.Errorf("save cloud connection verification: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil
	}
	if result.Status != app.CloudConnectionStatusVerified {
		signalCtx := context.WithValue(ctx, keys.AccountIDCtxKey, connection.CreatedByID)
		return a.enqueueOrgSignal(signalCtx, orgshelpers.EnqueueOrgSignalParams{
			OrgID: connection.OrgID,
			Signal: &verificationfailed.Signal{
				ConnectionID: connection.ID, ConnectionName: connection.Name, OrgID: connection.OrgID,
				Platform: connection.Platform, Message: result.Message,
			},
			IdempotencyKey: fmt.Sprintf("cloud-connection-verification-failed-%s-%s", connection.ID, now.Format(time.RFC3339Nano)),
		})
	}
	return nil
}
