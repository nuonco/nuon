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

func SyncPulumiComponent(ctx context.Context, db *gorm.DB, vcsHelper *vcshelpers.Helpers, comp *config.Component, componentID, appID, appConfigID string) (string, string, error) {
	if comp.Pulumi == nil {
		return "", "", sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: "pulumi config is required",
		}
	}

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

	obj := comp.Pulumi

	var githubVCSConfig *app.ConnectedGithubVCSConfig
	var publicGitConfig *app.PublicGitVCSConfig
	var err error

	if obj.ConnectedRepo != nil {
		githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
			Repo:      obj.ConnectedRepo.Repo,
			Branch:    obj.ConnectedRepo.Branch,
			Directory: obj.ConnectedRepo.Directory,
		}, parentCmp.App.Org)
		if err != nil {
			return "", "", sync.SyncInternalErr{
				Description: "unable to create connected github vcs config",
				Err:         fmt.Errorf("unable to create connected github vcs config: %w", err),
			}
		}
	}
	if obj.PublicRepo != nil {
		publicGitConfig = &app.PublicGitVCSConfig{
			Repo:      obj.PublicRepo.Repo,
			Branch:    obj.PublicRepo.Branch,
			Directory: obj.PublicRepo.Directory,
		}
	}

	pulumiCfg := app.PulumiComponentConfig{
		Runtime:                  obj.Runtime,
		Version:                  obj.PulumiVersion,
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
		Config:                   pgtype.Hstore{},
		EnvVars:                  pgtype.Hstore{},
	}

	for k, v := range obj.ConfigMap {
		v := v
		pulumiCfg.Config[k] = &v
	}
	for k, v := range obj.EnvVarMap {
		v := v
		pulumiCfg.EnvVars[k] = &v
	}

	var operationRoles pgtype.Hstore
	if len(comp.OperationRoles) > 0 {
		operationRoles = make(pgtype.Hstore)
		for _, opRole := range comp.OperationRoles {
			role := opRole.RoleName
			operationRoles[string(opRole.Operation)] = &role
		}
	}

	var references pq.StringArray
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	ccc := app.ComponentConfigConnection{
		PulumiComponentConfig: &pulumiCfg,
		ComponentID:           componentID,
		AppConfigID:           appConfigID,
		References:            references,
		BuildTimeout:          obj.BuildTimeout,
		DeployTimeout:         obj.DeployTimeout,
		MaxAutoRetries:        obj.MaxAutoRetries,
		SkipNoops:             obj.SkipNoops,
		OperationRoles:        operationRoles,
	}

	if obj.AutoApproveOnPoliciesPassing != nil {
		ccc.AutoApproveOnPoliciesPassing = obj.AutoApproveOnPoliciesPassing
	}
	if obj.DriftSchedule != nil {
		ccc.DriftSchedule = *obj.DriftSchedule
	}

	if dbErr := db.WithContext(ctx).Create(&ccc).Error; dbErr != nil {
		return "", "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create pulumi config for %s", comp.Name),
			Err:         dbErr,
		}
	}

	return ccc.ID, "", nil
}
