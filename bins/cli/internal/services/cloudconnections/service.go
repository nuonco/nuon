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

func (s *Service) Create(ctx context.Context, request *models.ServiceCreateRequest, asJSON bool) error {
	connection, err := s.api.CreateCloudConnection(ctx, request)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		return render(connection, true)
	}
	if err := render(connection, false); err != nil {
		return err
	}
	if request.Platform == "gcp" {
		ui.PrintLn("\nNext steps: configure GCP using the gcloud CLI or Terraform setup instructions:")
	} else if request.Platform == "azure" {
		ui.PrintLn("\nNext steps: configure Azure using the Azure CLI or Terraform setup instructions:")
	} else {
		ui.PrintLn("\nNext steps: configure AWS using the AWS CLI, Terraform, or CloudFormation setup instructions:")
	}
	cliConfig, err := s.api.GetCLIConfig(ctx)
	if err == nil && cliConfig.DashboardURL != "" {
		ui.Printf("  %s/%s/settings/cloud-connections/%s/setup\n", strings.TrimRight(cliConfig.DashboardURL, "/"), connection.OrgID, connection.ID)
	} else {
		ui.PrintLn("  Open your dashboard: Settings > Cloud connections > select this connection > View setup runbook.")
	}
	ui.Printf("\nAfter applying the cloud setup, verify the connection:\n  nuon cloud-connections verify %s\n", connection.ID)
	if connection.Preset == models.AppCloudConnectionPresetEcr || connection.Preset == models.AppCloudConnectionPresetGar || connection.Preset == models.AppCloudConnectionPresetAcr {
		ui.Printf("\nThen assign the image repositories it pulls:\n  nuon image-repositories create --connection-id %s --repository <repository>\n", connection.ID)
	}
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
	rows := [][]string{
		{"id", connection.ID}, {"name", connection.Name}, {"cloud", string(connection.Platform)},
		{"target", connection.TargetID}, {"principal", connection.Principal},
		{"status", string(connection.Status)}, {"last verified", connection.LastVerifiedAt}, {"preset", string(connection.Preset)},
		{"verification", verification},
		{"issuer", connection.Setup.IssuerURL}, {"subject", connection.Setup.Subject},
	}
	if connection.DefaultRegion != "" {
		rows = append(rows, []string{"region / location", connection.DefaultRegion})
	}
	if connection.GcpProjectNumber != "" {
		rows = append(rows, []string{"GCP project number", connection.GcpProjectNumber})
	}
	if connection.AzureRegistryID != "" {
		rows = append(rows, []string{"Azure tenant", connection.AzureTenantID}, []string{"ACR registry", connection.AzureRegistryID}, []string{"ACR login server", connection.AzureRegistryLoginServer})
	}
	ui.NewGetView().Render(rows)
	return nil
}

func (s *Service) ListRepositories(ctx context.Context, query, platform string, offset, limit int, asJSON bool) error {
	repositories, hasMore, err := s.api.ListImageRepositories(ctx, &nuon.ListImageRepositoriesQuery{
		Pagination: models.GetPaginatedQuery{Offset: offset, Limit: limit}, Q: query, Platform: platform,
	})
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(repositories)
		return nil
	}
	rows := [][]string{{"ID", "IMAGE URL", "CLOUD", "CONNECTION", "CONNECTION STATUS"}}
	for _, repository := range repositories {
		status := ""
		if repository.CloudConnection != nil {
			status = string(repository.CloudConnection.Status)
		}
		rows = append(rows, []string{repository.ID, repository.ImageURLPrefix, repository.Platform, repository.CloudConnectionID, status})
	}
	ui.NewListView().RenderPaging(rows, offset, limit, hasMore)
	return nil
}

func (s *Service) GetRepository(ctx context.Context, repositoryID string, asJSON bool) error {
	repository, err := s.api.GetImageRepository(ctx, repositoryID)
	if err != nil {
		return ui.PrintError(err)
	}
	return s.renderRepository(ctx, repository, asJSON)
}

func (s *Service) CreateRepository(ctx context.Context, request *models.ServiceCreateImageRepositoryRequest, asJSON bool) error {
	repository, err := s.api.CreateImageRepository(ctx, request)
	if err != nil {
		return ui.PrintError(err)
	}
	return s.renderRepository(ctx, repository, asJSON)
}

func (s *Service) DeleteRepository(ctx context.Context, repositoryID string, asJSON bool) error {
	if err := s.api.DeleteImageRepository(ctx, repositoryID); err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(map[string]any{"id": repositoryID, "deleted": true})
		return nil
	}
	ui.NewDeleteView("image repository", repositoryID, s.cfg.Interactive).Success()
	return nil
}

func (s *Service) renderRepository(ctx context.Context, repository *models.AppImageRepository, asJSON bool) error {
	if asJSON {
		ui.PrintJSON(repository)
		return nil
	}
	rows := [][]string{
		{"id", repository.ID}, {"image url", repository.ImageURLPrefix}, {"repository", repository.Repository},
		{"cloud", repository.Platform}, {"connection", repository.CloudConnectionID},
	}
	if repository.CloudConnection != nil {
		rows = append(rows, []string{"connection name", repository.CloudConnection.Name}, []string{"connection status", string(repository.CloudConnection.Status)})
	}
	ui.NewGetView().Render(rows)
	setup, err := s.api.GetCloudConnectionSetup(ctx, repository.CloudConnectionID, []string{repository.Repository})
	if err != nil {
		return ui.PrintError(err)
	}
	if setup.PermissionsCli != "" {
		ui.PrintLn("\nGrant the connection read access to this repository:")
		ui.PrintLn(setup.PermissionsCli)
	}
	if repository.CloudConnection != nil && repository.CloudConnection.Status != models.AppCloudConnectionStatusVerified {
		ui.Printf("\nThe connection is not verified yet. Finish its setup, then run:\n  nuon cloud-connections verify %s\n", repository.CloudConnectionID)
	}
	return nil
}
