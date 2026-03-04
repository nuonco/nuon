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

// SyncTerraformModuleComponent creates or updates a Terraform module component configuration.
// Duplicates logic from services/ctl-api/internal/app/components/service/create_terraform_module_component_config.go
func SyncTerraformModuleComponent(ctx context.Context, db *gorm.DB, comp *config.Component, componentID, appID, appConfigID string) (string, string, error) {
	// Validate Terraform component
	if err := validateTerraformComponent(comp); err != nil {
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
	vcsHelper := vcshelpers.New(vcshelpers.Params{})

	var githubVCSConfig *app.ConnectedGithubVCSConfig
	var publicGitConfig *app.PublicGitVCSConfig
	var err error

	if comp.TerraformModule != nil {
		if comp.TerraformModule.ConnectedRepo != nil {
			githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
				Repo:      comp.TerraformModule.ConnectedRepo.Repo,
				Branch:    comp.TerraformModule.ConnectedRepo.Branch,
				Directory: comp.TerraformModule.ConnectedRepo.Directory,
			}, parentCmp.App.Org)
			if err != nil {
				return "", "", sync.SyncInternalErr{
					Description: "unable to create connected github vcs config",
					Err:         fmt.Errorf("unable to create connected github vcs config: %w", err),
				}
			}
		}

		if comp.TerraformModule.PublicRepo != nil {
			publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
				Repo:      comp.TerraformModule.PublicRepo.Repo,
				Branch:    comp.TerraformModule.PublicRepo.Branch,
				Directory: comp.TerraformModule.PublicRepo.Directory,
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
	}

	// Build variables map
	variables := make(map[string]*string)
	if comp.TerraformModule.VarsMap != nil {
		for k, v := range comp.TerraformModule.VarsMap {
			val := v
			variables[k] = &val
		}
	}

	// Build env vars map
	envVars := make(map[string]*string)
	if comp.TerraformModule.EnvVarMap != nil {
		for k, v := range comp.TerraformModule.EnvVarMap {
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

	// Get variables files
	variablesFiles := []string{}
	if comp.TerraformModule.VariablesFiles != nil {
		for _, vf := range comp.TerraformModule.VariablesFiles {
			// TerraformVariablesFile only has Contents field
			if vf.Contents != "" {
				variablesFiles = append(variablesFiles, vf.Contents)
			}
		}
	}

	// Get version (default to v1.7.5 if not provided - handled by GORM default tag)
	version := comp.TerraformModule.TerraformVersion
	if version == "" {
		version = "v1.7.5" // This matches the GORM default, but ideally should get latest
	}

	// Create Terraform module component config
	cfg := app.TerraformModuleComponentConfig{
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
		Version:                  version,
		Variables:                pgtype.Hstore(variables),
		EnvVars:                  pgtype.Hstore(envVars),
		VariablesFiles:           pq.StringArray(variablesFiles),
	}

	// Get references
	references := []string{}
	if comp.References != nil {
		for _, ref := range comp.References {
			references = append(references, ref.String())
		}
	}

	// Get drift schedule if present
	driftSchedule := ""
	if comp.TerraformModule.DriftSchedule != nil {
		driftSchedule = *comp.TerraformModule.DriftSchedule
	}

	// Create component config connection
	componentConfigConnection := app.ComponentConfigConnection{
		TerraformModuleComponentConfig: &cfg,
		ComponentID:                    componentID,
		AppConfigID:                    appConfigID,
		ComponentDependencyIDs:         pq.StringArray(depIDs),
		References:                     pq.StringArray(references),
		Checksum:                       comp.Checksum,
		BuildTimeout:                   comp.TerraformModule.BuildTimeout,
		DeployTimeout:                  comp.TerraformModule.DeployTimeout,
		OperationRoles:                 operationRoles,
		DriftSchedule:                  driftSchedule,
	}

	res = db.WithContext(ctx).Create(&componentConfigConnection)
	if res.Error != nil {
		return "", "", sync.SyncInternalErr{
			Description: "unable to create terraform module component config",
			Err:         res.Error,
		}
	}

	return componentConfigConnection.ID, componentConfigConnection.Checksum, nil
}

// validateTerraformComponent validates Terraform module component configuration.
// Duplicates validation logic from services/ctl-api/internal/app/components/service/create_terraform_module_component_config.go
func validateTerraformComponent(comp *config.Component) error {
	if comp.TerraformModule == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("terraform module config is required"),
			Description: fmt.Sprintf("Component '%s' is missing terraform module configuration", comp.Name),
		}
	}

	// Validate VCS configuration (required for Terraform)
	if comp.TerraformModule.ConnectedRepo == nil && comp.TerraformModule.PublicRepo == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("vcs_config_required"),
			Code:        "vcs_config_required",
			Description: fmt.Sprintf("Component '%s' requires either a connected repo or public repo configuration", comp.Name),
		}
	}

	// Validate Terraform version if provided
	// Note: In the API, if version is empty it defaults to latest from tfClient.GetLatestVersion()
	// For now, we'll allow empty and let GORM default handle it
	if comp.TerraformModule.TerraformVersion != "" {
		// Validate version is within min/max bounds
		// TODO: Get latest version dynamically from tfClient.GetLatestVersion()
		// For now, using hardcoded values that match the API
		minVersion := "1.8.0"
		maxVersion := "1.10.5" // This should come from tfClient
		if err := validation.ValidateTerraformVersion(comp.TerraformModule.TerraformVersion, minVersion, maxVersion); err != nil {
			return err
		}
	}

	// Validate timeouts if provided
	if comp.TerraformModule.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(comp.TerraformModule.BuildTimeout); err != nil {
			return err
		}
	}

	if comp.TerraformModule.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(comp.TerraformModule.DeployTimeout); err != nil {
			return err
		}
	}

	// Validate drift schedule if provided
	if comp.TerraformModule.DriftSchedule != nil && *comp.TerraformModule.DriftSchedule != "" {
		if err := validation.ValidateCronSchedule(*comp.TerraformModule.DriftSchedule); err != nil {
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
