package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/signals/verificationfailed"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type stsAPI interface {
	AssumeRoleWithWebIdentity(context.Context, *sts.AssumeRoleWithWebIdentityInput, ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

func (h *Helpers) assumeRole(ctx context.Context, connection *app.CloudConnection, client stsAPI, input *sts.AssumeRoleWithWebIdentityInput) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	output, err := client.AssumeRoleWithWebIdentity(ctx, input)
	if !cloudconnections.IsAccessDenied(err) {
		return output, err
	}
	now := time.Now().UTC()
	message := cloudconnections.AssumeRoleErrorMessage(err)
	update := app.CloudConnection{Status: app.CloudConnectionStatusError, StatusMessage: message, LastVerifiedAt: &now}
	result := h.db.WithContext(ctx).Model(&app.CloudConnection{}).
		Where(app.CloudConnection{OrgID: connection.OrgID, ID: connection.ID, Principal: connection.Principal}).
		Select("status", "status_message", "last_verified_at").Updates(update)
	if result.Error != nil {
		return nil, errors.Join(err, fmt.Errorf("save cloud connection verification: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, err
	}
	signalCtx := context.WithValue(ctx, keys.AccountIDCtxKey, connection.CreatedByID)
	if signalErr := h.enqueueOrgSignal(signalCtx, orgshelpers.EnqueueOrgSignalParams{
		OrgID: connection.OrgID,
		Signal: &verificationfailed.Signal{
			ConnectionID: connection.ID, ConnectionName: connection.Name, OrgID: connection.OrgID,
			Platform: connection.Platform, Message: message,
		},
		IdempotencyKey: fmt.Sprintf("cloud-connection-verification-failed-%s-%s", connection.ID, now.Format(time.RFC3339Nano)),
	}); signalErr != nil {
		return nil, errors.Join(err, fmt.Errorf("enqueue cloud connection verification failure: %w", signalErr))
	}
	return nil, err
}
