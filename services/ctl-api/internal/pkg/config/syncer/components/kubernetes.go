package components

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

// SyncKubernetesManifestComponent creates or updates a Kubernetes manifest component configuration.
func SyncKubernetesManifestComponent(ctx context.Context, db *gorm.DB, comp *config.Component, componentID, appID, appConfigID string) (string, string, error) {
	if comp.KubernetesManifest == nil {
		return "", "", sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: "kubernetes manifest config is required",
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

	// Build VCS configs
	vcsHelper := vcshelpers.New(vcshelpers.Params{})

	var githubVCSConfig *app.ConnectedGithubVCSConfig
	var publicGitConfig *app.PublicGitVCSConfig
	var err error

	if comp.KubernetesManifest.ConnectedRepo != nil {
		githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
			Repo:      comp.KubernetesManifest.ConnectedRepo.Repo,
			Branch:    comp.KubernetesManifest.ConnectedRepo.Branch,
			Directory: comp.KubernetesManifest.ConnectedRepo.Directory,
		}, parentCmp.App.Org)
		if err != nil {
			return "", "", sync.SyncInternalErr{
				Description: "unable to create connected github vcs config",
				Err:         err,
			}
		}
	}

	if comp.KubernetesManifest.PublicRepo != nil {
		publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
			Repo:      comp.KubernetesManifest.PublicRepo.Repo,
			Branch:    comp.KubernetesManifest.PublicRepo.Branch,
			Directory: comp.KubernetesManifest.PublicRepo.Directory,
		})
		if err != nil {
			return "", "", sync.SyncInternalErr{
				Description: "unable to get public git config",
				Err:         err,
			}
		}
	}

	// Build kustomize config if present
	var kustomize *app.KustomizeConfig
	if comp.KubernetesManifest.Kustomize != nil {
		kustomize = &app.KustomizeConfig{
			Path: comp.KubernetesManifest.Kustomize.Path,
		}
	}

	// Build operation roles
	operationRoles := make(pgtype.Hstore)
	if comp.OperationRoles != nil {
		for _, role := range comp.OperationRoles {
			operationRoles[string(role.Operation)] = &role.RoleName
		}
	}

	// Create kubernetes manifest component config
	cfg := app.KubernetesManifestComponentConfig{
		Namespace:                comp.KubernetesManifest.Namespace,
		Kustomize:                kustomize,
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
	}

	// Build references
	references := []string{}
	if comp.References != nil {
		for _, ref := range comp.References {
			references = append(references, ref.String())
		}
	}

	// Create component config connection
	componentConfigConnection := app.ComponentConfigConnection{
		KubernetesManifestComponentConfig: &cfg,
		ComponentID:                       componentID,
		AppConfigID:                       appConfigID,
		ComponentDependencyIDs:            pq.StringArray{},
		References:                        pq.StringArray(references),
		Checksum:                          comp.Checksum,
		OperationRoles:                    operationRoles,
	}

	res = db.WithContext(ctx).Create(&componentConfigConnection)
	if res.Error != nil {
		return "", "", sync.SyncInternalErr{
			Description: "unable to create kubernetes manifest component config",
			Err:         res.Error,
		}
	}

	return componentConfigConnection.ID, componentConfigConnection.Checksum, nil
}
