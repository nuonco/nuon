package cloudconnections

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (s *Service) List(ctx context.Context, offset, limit int, asJSON bool) error {
	connections, hasMore, err := s.api.ListCloudConnections(ctx, &models.GetPaginatedQuery{Offset: offset, Limit: limit})
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
	ui.NewListView().RenderPaging(rows, offset, limit, hasMore)
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
	if asJSON {
		return render(connection, true)
	}
	if err := render(connection, false); err != nil {
		return err
	}
	ui.PrintLn("\nNext steps: configure AWS using the AWS CLI, Terraform, or CloudFormation setup instructions:")
	cliConfig, err := s.api.GetCLIConfig(ctx)
	if err == nil && cliConfig.DashboardURL != "" {
		ui.Printf("  %s/%s/settings/cloud-connections/%s/setup\n", strings.TrimRight(cliConfig.DashboardURL, "/"), connection.OrgID, connection.ID)
	} else {
		ui.PrintLn("  Open your dashboard: Settings > Cloud connections > select this connection > View setup runbook.")
	}
	ui.Printf("\nAfter applying the AWS setup, verify the connection:\n  nuon cloud-connections verify %s\n", connection.ID)
	return nil
}

func (s *Service) Verify(ctx context.Context, connectionID string, wait, asJSON bool) error {
	connection, err := s.api.VerifyCloudConnection(ctx, connectionID)
	if err != nil {
		return ui.PrintError(err)
	}
	if wait {
		pollCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		timeoutErr := &ui.CLIUserError{Msg: fmt.Sprintf("cloud connection %s verification timed out after 2 minutes", connectionID)}
		for connection.VerificationInProgress {
			select {
			case <-pollCtx.Done():
				if ctx.Err() != nil {
					return ui.PrintError(ctx.Err())
				}
				return ui.PrintError(timeoutErr)
			case <-ticker.C:
			}
			updated, err := s.api.GetCloudConnection(pollCtx, connectionID)
			if err != nil {
				if pollCtx.Err() != nil && ctx.Err() == nil {
					return ui.PrintError(timeoutErr)
				}
				return ui.PrintError(err)
			}
			connection = updated
		}
		if connection.Status != models.AppCloudConnectionStatusVerified {
			return ui.PrintError(&ui.CLIUserError{Msg: fmt.Sprintf("cloud connection %s verification ended with status %s: %s", connectionID, connection.Status, connection.StatusMessage)})
		}
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
	verification := connection.StatusMessage
	if connection.VerificationInProgress {
		verification = "Still checking — refresh in a moment"
	}
	ui.NewGetView().Render([][]string{
		{"id", connection.ID}, {"name", connection.Name}, {"cloud", string(connection.Platform)},
		{"target", connection.TargetID}, {"principal", connection.Principal},
		{"status", string(connection.Status)}, {"last verified", connection.LastVerifiedAt}, {"preset", string(connection.Preset)},
		{"verification", verification},
		{"issuer", connection.Setup.IssuerURL}, {"subject", connection.Setup.Subject},
	})
	return nil
}
