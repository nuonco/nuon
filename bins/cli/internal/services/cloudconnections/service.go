package cloudconnections

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type Service struct {
	api nuon.Client
	cfg *config.Config
}

func New(apiClient nuon.Client, cfg *config.Config) *Service {
	return &Service{api: apiClient, cfg: cfg}
}

func (s *Service) List(ctx context.Context, asJSON bool) error {
	connections, err := s.api.ListCloudConnections(ctx)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(connections)
		return nil
	}
	rows := [][]string{{"ID", "NAME", "CLOUD", "TARGET", "PRINCIPAL", "STATUS", "LAST VERIFIED", "REQUESTED", "VERIFIED"}}
	for _, connection := range connections {
		rows = append(rows, []string{connection.ID, connection.Name, string(connection.Platform), connection.TargetID, connection.Principal, string(connection.Status), connection.LastVerifiedAt, capabilitiesString(connection.RequestedCapabilities), capabilitiesString(connection.Capabilities)})
	}
	ui.NewListView().Render(rows)
	return nil
}

func (s *Service) Get(ctx context.Context, connectionID string, asJSON bool) error {
	connection, err := s.api.GetCloudConnection(ctx, connectionID)
	if err != nil {
		return ui.PrintError(err)
	}
	return render(connection, asJSON)
}

func (s *Service) Create(ctx context.Context, name, platform, targetID, principal, tenantID, identityProvider, defaultRegion, registry string, capabilities, repositories []string, asJSON bool) error {
	capabilityValues := make([]models.AppCloudConnectionCapability, 0, len(capabilities))
	for _, capability := range capabilities {
		capabilityValues = append(capabilityValues, models.AppCloudConnectionCapability(capability))
	}
	connection, err := s.api.CreateCloudConnection(ctx, &models.ServiceCreateRequest{
		Name: name, Platform: models.AppCloudPlatform(platform), TargetID: targetID,
		Principal: principal, TenantID: tenantID, IdentityProvider: identityProvider, DefaultRegion: defaultRegion,
		Capabilities: capabilityValues, Registry: registry, Repositories: repositories,
	})
	if err != nil {
		return ui.PrintError(err)
	}
	return render(connection, asJSON)
}

func (s *Service) Verify(ctx context.Context, connectionID, registry string, repositories []string, asJSON bool) error {
	connection, err := s.api.VerifyCloudConnection(ctx, connectionID, &models.ServiceVerifyRequest{Registry: registry, Repositories: repositories})
	if err != nil {
		return ui.PrintError(err)
	}
	return render(connection, asJSON)
}

func (s *Service) Delete(ctx context.Context, connectionID string, asJSON bool) error {
	if err := s.api.DeleteCloudConnection(ctx, connectionID); err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(map[string]any{"id": connectionID, "deleted": true})
		return nil
	}
	ui.NewDeleteView("cloud connection", connectionID, s.cfg.Interactive).Success()
	return nil
}

func render(connection *models.ServiceConnectionResponse, asJSON bool) error {
	if asJSON {
		ui.PrintJSON(connection)
		return nil
	}
	ui.NewGetView().Render([][]string{
		{"id", connection.ID}, {"name", connection.Name}, {"cloud", string(connection.Platform)},
		{"target", connection.TargetID}, {"principal", connection.Principal},
		{"status", string(connection.Status)}, {"last verified", connection.LastVerifiedAt}, {"requested capabilities", capabilitiesString(connection.RequestedCapabilities)}, {"verified capabilities", capabilitiesString(connection.Capabilities)},
		{"issuer", connection.Setup.IssuerURL}, {"subject", connection.Setup.Subject},
	})
	return nil
}

func capabilitiesString(capabilities []models.AppCloudConnectionCapability) string {
	values := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		values = append(values, string(capability))
	}
	return strings.Join(values, ", ")
}
