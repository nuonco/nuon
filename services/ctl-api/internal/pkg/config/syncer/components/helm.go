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

// SyncHelmComponent creates or updates a Helm component configuration.
// Duplicates logic from services/ctl-api/internal/app/components/service/create_helm_component_config.go
func SyncHelmComponent(ctx context.Context, db *gorm.DB, comp *config.Component, componentID, appID, appConfigID string) (string, string, error) {
	// Validate Helm component
	if err := validateHelmComponent(comp); err != nil {
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

	// Build VCS configs using helpers (optional for Helm with repo config)
	vcsHelper := vcshelpers.New(vcshelpers.Params{})

	var githubVCSConfig *app.ConnectedGithubVCSConfig
	var publicGitConfig *app.PublicGitVCSConfig
	var err error

	if comp.HelmChart != nil {
		if comp.HelmChart.ConnectedRepo != nil {
			githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
				Repo:      comp.HelmChart.ConnectedRepo.Repo,
				Branch:    comp.HelmChart.ConnectedRepo.Branch,
				Directory: comp.HelmChart.ConnectedRepo.Directory,
			}, parentCmp.App.Org)
			if err != nil {
				return "", "", sync.SyncInternalErr{
					Description: "unable to create connected github vcs config",
					Err:         fmt.Errorf("unable to create connected github vcs config: %w", err),
				}
			}
		}

		if comp.HelmChart.PublicRepo != nil {
			publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
				Repo:      comp.HelmChart.PublicRepo.Repo,
				Branch:    comp.HelmChart.PublicRepo.Branch,
				Directory: comp.HelmChart.PublicRepo.Directory,
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

	// Build values map
	values := make(map[string]*string)
	if comp.HelmChart.ValuesMap != nil {
		for k, v := range comp.HelmChart.ValuesMap {
			val := v
			values[k] = &val
		}
	}

	// Build operation roles map
	operationRoles := make(pgtype.Hstore)
	if comp.OperationRoles != nil {
		for _, role := range comp.OperationRoles {
			operationRoles[string(role.Operation)] = &role.RoleName
		}
	}

	// Build Helm repo config if present
	var helmRepoConfig *app.HelmRepoConfig
	if comp.HelmChart.HelmRepo != nil {
		helmRepoConfig = &app.HelmRepoConfig{
			RepoURL: comp.HelmChart.HelmRepo.RepoURL,
			Chart:   comp.HelmChart.HelmRepo.Chart,
			Version: comp.HelmChart.HelmRepo.Version,
		}
	}

	// Convert values files
	var valuesFiles []string
	for _, vf := range comp.HelmChart.ValuesFiles {
		// Use Path if available, otherwise Contents
		if vf.Path != "" {
			valuesFiles = append(valuesFiles, vf.Path)
		} else if vf.Contents != "" {
			valuesFiles = append(valuesFiles, vf.Contents)
		}
	}

	// Create Helm config
	helmConfig := &app.HelmConfig{
		ChartName:      comp.HelmChart.ChartName,
		Namespace:      comp.HelmChart.Namespace,
		Values:         values,
		ValuesFiles:    valuesFiles,
		StorageDriver:  comp.HelmChart.StorageDriver,
		TakeOwnership:  comp.HelmChart.TakeOwnership,
		HelmRepoConfig: helmRepoConfig,
	}

	// Create Helm component config
	cfg := app.HelmComponentConfig{
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
		HelmConfig:               helmConfig,
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
	if comp.HelmChart.DriftSchedule != nil {
		driftSchedule = *comp.HelmChart.DriftSchedule
	}

	// Create component config connection
	componentConfigConnection := app.ComponentConfigConnection{
		HelmComponentConfig:    &cfg,
		ComponentID:            componentID,
		AppConfigID:            appConfigID,
		ComponentDependencyIDs: pq.StringArray(depIDs),
		References:             pq.StringArray(references),
		Checksum:               comp.Checksum,
		BuildTimeout:           comp.HelmChart.BuildTimeout,
		DeployTimeout:          comp.HelmChart.DeployTimeout,
		OperationRoles:         operationRoles,
		DriftSchedule:          driftSchedule,
	}

	res = db.WithContext(ctx).Create(&componentConfigConnection)
	if res.Error != nil {
		return "", "", sync.SyncInternalErr{
			Description: "unable to create helm component config",
			Err:         res.Error,
		}
	}

	return componentConfigConnection.ID, componentConfigConnection.Checksum, nil
}

// validateHelmComponent validates Helm component configuration.
// Duplicates validation logic from services/ctl-api/internal/app/components/service/create_helm_component_config.go
func validateHelmComponent(comp *config.Component) error {
	if comp.HelmChart == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("helm chart config is required"),
			Description: fmt.Sprintf("Component '%s' is missing helm chart configuration", comp.Name),
		}
	}

	// Validate VCS configuration OR helm repo config (one must be present)
	hasVCSConfig := comp.HelmChart.ConnectedRepo != nil || comp.HelmChart.PublicRepo != nil
	hasHelmRepoConfig := comp.HelmChart.HelmRepo != nil

	if !hasVCSConfig && !hasHelmRepoConfig {
		return stderr.ErrUser{
			Err:         fmt.Errorf("vcs_config_required"),
			Code:        "vcs_config_required",
			Description: fmt.Sprintf("Component '%s' requires either a VCS configuration or Helm repository configuration", comp.Name),
		}
	}

	// Validate chart name (required, DNS RFC 1035 label, 5-62 chars)
	if comp.HelmChart.ChartName == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("chart_name is required"),
			Description: fmt.Sprintf("Component '%s' is missing required field 'chart_name'", comp.Name),
		}
	}

	if err := validation.ValidateDNSName(comp.HelmChart.ChartName, 5, 62); err != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("invalid chart_name: %s", comp.HelmChart.ChartName),
			Description: fmt.Sprintf("Component '%s' has invalid chart_name '%s'. Must be a valid DNS RFC 1035 label (5-62 characters, start with letter, lowercase letters/numbers/hyphens only)", comp.Name, comp.HelmChart.ChartName),
		}
	}

	// Validate timeouts if provided
	if comp.HelmChart.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(comp.HelmChart.BuildTimeout); err != nil {
			return err
		}
	}

	if comp.HelmChart.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(comp.HelmChart.DeployTimeout); err != nil {
			return err
		}
	}

	// Validate drift schedule if provided
	if comp.HelmChart.DriftSchedule != nil && *comp.HelmChart.DriftSchedule != "" {
		if err := validation.ValidateCronSchedule(*comp.HelmChart.DriftSchedule); err != nil {
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
