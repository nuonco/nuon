package components

import (
	"context"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/validation"
)

// SyncDockerBuildComponent creates or updates a Docker build component configuration.
// Duplicates logic from services/ctl-api/internal/app/components/service/create_docker_build_component_config.go
func SyncDockerBuildComponent(ctx context.Context, db *gorm.DB, vcsHelper *vcshelpers.Helpers, comp *config.Component, componentID, appID, appConfigID string) (string, string, error) {
	// Validate Docker build component
	if err := validateDockerBuildComponent(comp); err != nil {
		return "", "", sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: fmt.Sprintf("validation failed: %v", err),
		}
	}

	// Get the component with app and org preloaded for VCS helpers
	var parentCmp app.Component
	res := db.WithContext(ctx).
		Preload("App").
		Preload("App.Org").
		Preload("App.Org.VCSConnections").
		First(&parentCmp, "id = ?", componentID)
	if res.Error != nil {
		return "", "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to get component %s", comp.Name),
			Err:         res.Error,
		}
	}

	// Build VCS configs using helpers

	var githubVCSConfig *app.ConnectedGithubVCSConfig
	var publicGitConfig *app.PublicGitVCSConfig
	var err error

	if comp.DockerBuild != nil {
		if comp.DockerBuild.ConnectedRepo != nil {
			githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
				Repo:      comp.DockerBuild.ConnectedRepo.Repo,
				Branch:    comp.DockerBuild.ConnectedRepo.Branch,
				Directory: comp.DockerBuild.ConnectedRepo.Directory,
			}, parentCmp.App.Org)
			if err != nil {
				return "", "", sync.SyncInternalErr{
					Description: "unable to create connected github vcs config",
					Err:         fmt.Errorf("unable to create connected github vcs config: %w", err),
				}
			}
		}

		if comp.DockerBuild.PublicRepo != nil {
			publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
				Repo:      comp.DockerBuild.PublicRepo.Repo,
				Branch:    comp.DockerBuild.PublicRepo.Branch,
				Directory: comp.DockerBuild.PublicRepo.Directory,
			})
			if err != nil {
				return "", "", sync.SyncInternalErr{
					Description: "unable to get public git config",
					Err:         fmt.Errorf("unable to get public git config: %w", err),
				}
			}
		}
	}

	// Resolve component dependencies
	depIDs := []string{}
	if len(comp.Dependencies) > 0 {
		// TODO: Implement GetComponentIDs helper
		// For now, we'll leave dependencies empty
		// depIDs, err = helpers.GetComponentIDs(ctx, appID, comp.Dependencies)
	}

	// Build env vars map
	envVars := make(map[string]*string)
	if comp.DockerBuild != nil && comp.DockerBuild.EnvVarMap != nil {
		for k, v := range comp.DockerBuild.EnvVarMap {
			val := v
			envVars[k] = &val
		}
	}

	// Build operation roles map
	operationRoles := make(pgtype.Hstore)
	if comp.OperationRoles != nil {
		for _, role := range comp.OperationRoles {
			operationRoles[string(role.Operation)] = &role.RoleName
		}
	}

	// Get dockerfile (default to "Dockerfile")
	dockerfile := "Dockerfile"
	if comp.DockerBuild != nil && comp.DockerBuild.Dockerfile != "" {
		dockerfile = comp.DockerBuild.Dockerfile
	}

	// Create Docker build component config
	cfg := app.DockerBuildComponentConfig{
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
		Dockerfile:               dockerfile,
		EnvVars:                  pgtype.Hstore(envVars),
	}

	// Get references
	references := []string{}
	if comp.References != nil {
		for _, ref := range comp.References {
			references = append(references, ref.String())
		}
	}

	// Create component config connection
	componentConfigConnection := app.ComponentConfigConnection{
		DockerBuildComponentConfig: &cfg,
		ComponentID:                componentID,
		AppConfigID:                appConfigID,
		ComponentDependencyIDs:     pq.StringArray(depIDs),
		References:                 pq.StringArray(references),
		Checksum:                   comp.Checksum,
		BuildTimeout:               comp.DockerBuild.BuildTimeout,
		DeployTimeout:              comp.DockerBuild.DeployTimeout,
		MaxAutoRetries:             comp.DockerBuild.MaxAutoRetries,
		OperationRoles:             operationRoles,
	}

	res = db.WithContext(ctx).Create(&componentConfigConnection)
	if res.Error != nil {
		return "", "", sync.SyncInternalErr{
			Description: "unable to create docker build component config",
			Err:         res.Error,
		}
	}

	return componentConfigConnection.ID, componentConfigConnection.Checksum, nil
}

// validateDockerBuildComponent validates Docker build component configuration.
// Duplicates validation logic from services/ctl-api/internal/app/components/service/create_docker_build_component_config.go
func validateDockerBuildComponent(comp *config.Component) error {
	if comp.DockerBuild == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("docker build config is required"),
			Description: fmt.Sprintf("Component '%s' is missing docker build configuration", comp.Name),
		}
	}

	// Validate VCS configuration (one must be present)
	if comp.DockerBuild.ConnectedRepo == nil && comp.DockerBuild.PublicRepo == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("vcs_config_required"),
			Code:        "vcs_config_required",
			Description: fmt.Sprintf("Component '%s' requires either a connected repo or public repo configuration", comp.Name),
		}
	}

	// Validate timeouts if provided
	if comp.DockerBuild.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(comp.DockerBuild.BuildTimeout); err != nil {
			return err
		}
	}

	if comp.DockerBuild.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(comp.DockerBuild.DeployTimeout); err != nil {
			return err
		}
	}

	if comp.DockerBuild.MaxAutoRetries != nil {
		if err := validation.ValidateMaxAutoRetries(*comp.DockerBuild.MaxAutoRetries); err != nil {
			return err
		}
	}

	// Validate operation roles if provided
	if comp.OperationRoles != nil {
		for _, role := range comp.OperationRoles {
			if !slices.Contains(app.ValidOperations, app.OperationType(role.Operation)) {
				return stderr.ErrUser{
					Err:         fmt.Errorf("invalid operation type: %s", role.Operation),
					Description: fmt.Sprintf("Component '%s' has invalid operation type '%s'. Valid operations: %v", comp.Name, role.Operation, app.ValidOperations),
				}
			}
		}
	}

	return nil
}
