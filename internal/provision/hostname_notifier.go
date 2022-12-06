package provision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-sender"
)

type SendHostnameNotificationRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	InstallID string `json:"install_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
}

func (s SendHostnameNotificationRequest) validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type SendHostnameNotificationResponse struct{}

func (a *Activities) SendHostnameNotification(ctx context.Context, req SendHostnameNotificationRequest) (SendHostnameNotificationResponse, error) {
	var resp SendHostnameNotificationResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	if err := a.sendHostnameNotification(ctx, client, req); err != nil {
		return resp, fmt.Errorf("unable to send hostname notification: %w", err)
	}

	return resp, nil
}

type hostnameNotificationSender interface {
	sendHostnameNotification(context.Context, waypointClientHostnameGetter, SendHostnameNotificationRequest) error
}

var _ hostnameNotificationSender = (*hostnameNotificationSenderImpl)(nil)

type hostnameNotificationSenderImpl struct {
	//nolint:all
	sender sender.NotificationSender
}

type waypointClientHostnameGetter interface{}

func (h *hostnameNotificationSenderImpl) sendHostnameNotification(ctx context.Context, client waypointClientHostnameGetter, req SendHostnameNotificationRequest) error {
	return nil
}
