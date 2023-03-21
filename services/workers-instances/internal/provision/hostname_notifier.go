package provision

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/sender"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"google.golang.org/grpc"
)

var (
	errNoHostnamesFound error = fmt.Errorf("no hostnames found")
)

const (
	hostnameNotificationTemplate = `:package: _successfully provisioned deployment_:package:
• *install-id*: %s
• *org-id*: _%s_
• *url*: _%s_
`
)

type SendHostnameNotificationRequest struct {
	OrgID                string `json:"org_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	InstallID   string `json:"install_id" validate:"required"`
	ComponentID string `json:"component_id" validate:"required"`
}

func (s SendHostnameNotificationRequest) validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type SendHostnameNotificationResponse struct {
	Hostname string `json:"hostname" validate:"required"`
}

func (a *Activities) SendHostnameNotification(ctx context.Context, req SendHostnameNotificationRequest) (SendHostnameNotificationResponse, error) {
	var resp SendHostnameNotificationResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	provider, err := waypoint.NewOrgProvider(a.v, waypoint.WithOrgConfig(waypoint.Config{
		Address: req.OrgServerAddr,
		Token: waypoint.Token{
			Namespace: req.TokenSecretNamespace,
			Name:      waypoint.DefaultTokenSecretName(req.OrgID),
		},
	}))
	if err != nil {
		return resp, fmt.Errorf("unable to get org provider: %w", err)
	}

	client, err := provider.GetClient(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
	}

	hostname, err := a.getHostname(ctx, client, req)
	if err != nil {
		if errors.Is(errNoHostnamesFound, err) {
			return resp, nil
		}

		return resp, fmt.Errorf("unable to get hostname: %w", err)
	}
	resp.Hostname = hostname

	if err := a.sendHostnameNotification(ctx, hostname, req); err != nil {
		return resp, fmt.Errorf("unable to send hostname notification")
	}

	return resp, nil
}

type hostnameNotificationSender interface {
	getHostname(context.Context, waypointClientHostnameGetter, SendHostnameNotificationRequest) (string, error)
	sendHostnameNotification(context.Context, string, SendHostnameNotificationRequest) error
}

var _ hostnameNotificationSender = (*hostnameNotificationSenderImpl)(nil)

type hostnameNotificationSenderImpl struct {
	sender sender.NotificationSender
}

type waypointClientHostnameGetter interface {
	ListHostnames(context.Context, *gen.ListHostnamesRequest, ...grpc.CallOption) (*gen.ListHostnamesResponse, error)
}

func (h *hostnameNotificationSenderImpl) sendHostnameNotification(ctx context.Context, hostname string, req SendHostnameNotificationRequest) error {
	msg := fmt.Sprintf(hostnameNotificationTemplate, req.InstallID, req.OrgID, hostname)

	if err := h.sender.Send(ctx, msg); err != nil {
		return fmt.Errorf("unable to send notification: %w", err)
	}
	return nil
}

func (h *hostnameNotificationSenderImpl) getHostname(ctx context.Context, client waypointClientHostnameGetter, req SendHostnameNotificationRequest) (string, error) {
	wpReq := &gen.ListHostnamesRequest{
		Target: &gen.Hostname_Target{
			Target: &gen.Hostname_Target_Application{
				Application: &gen.Hostname_TargetApp{
					Application: &gen.Ref_Application{
						Application: req.ComponentID,
						Project:     req.InstallID,
					},
					Workspace: &gen.Ref_Workspace{
						Workspace: req.InstallID,
					},
				},
			},
		},
	}
	resp, err := client.ListHostnames(ctx, wpReq)
	if err != nil {
		return "", fmt.Errorf("unable to list hostnames for app: %w", err)
	}

	if len(resp.Hostnames) < 1 {
		return "", errNoHostnamesFound
	}

	return resp.Hostnames[0].Fqdn, nil
}
