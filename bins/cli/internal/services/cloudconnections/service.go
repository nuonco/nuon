package cloudconnections

import (
	"context"

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
	rows := [][]string{{"ID", "NAME", "CLOUD", "TARGET", "PRINCIPAL", "STATUS", "LAST VERIFIED", "PRESET"}}
	for _, connection := range connections {
		rows = append(rows, []string{connection.ID, connection.Name, string(connection.Platform), connection.TargetID, connection.Principal, string(connection.Status), connection.LastVerifiedAt, string(connection.Preset)})
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

func (s *Service) Create(ctx context.Context, name, platform, targetID, principal, defaultRegion, preset string, asJSON bool) error {
	connection, err := s.api.CreateCloudConnection(ctx, &models.ServiceCreateRequest{
		Name: name, Platform: platform, TargetID: targetID,
		Principal: principal, DefaultRegion: defaultRegion, Preset: models.AppCloudConnectionPreset(preset),
	})
	if err != nil {
		return ui.PrintError(err)
	}
	return render(connection, asJSON)
}

func (s *Service) Verify(ctx context.Context, connectionID string, asJSON bool) error {
	connection, err := s.api.VerifyCloudConnection(ctx, connectionID, nil)
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
		{"status", string(connection.Status)}, {"last verified", connection.LastVerifiedAt}, {"preset", string(connection.Preset)},
		{"issuer", connection.Setup.IssuerURL}, {"subject", connection.Setup.Subject},
	})
	return nil
}
