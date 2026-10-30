package controlplanejob

import (
	"context"
	"fmt"
	"strings"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (a *Activities) authorizeGARSource(ctx context.Context, jobID string, plan *plantypes.CompositePlan) error {
	var source *configs.OCIRegistryRepository
	if plan.BuildPlan != nil && plan.BuildPlan.ContainerImagePullPlan != nil {
		source = plan.BuildPlan.ContainerImagePullPlan.RepoCfg
	}
	if plan.FetchImageMetadataPlan != nil {
		source = plan.FetchImageMetadataPlan.Registry
	}
	if source == nil || source.RegistryType != configs.OCIRegistryTypeGAR {
		return nil
	}
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.OrgID == "" || job.OwnerID == "" || job.OwnerType != "component_builds" {
		return fmt.Errorf("GAR source requires an org-owned component build")
	}
	var build app.ComponentBuild
	if err := a.db.WithContext(ctx).
		Where(app.ComponentBuild{ID: job.OwnerID, OrgID: job.OrgID}).
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.GCPGARImageConfig").
		First(&build).Error; err != nil {
		return fmt.Errorf("unable to get GAR source build: %w", err)
	}
	cfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if cfg == nil || cfg.GCPGARImageConfig == nil || source.Repository != cfg.ImageURL {
		return fmt.Errorf("component build has no matching GAR source configuration")
	}
	token, err := a.cloudConnectionsHelpers.GARAuthorization(ctx, build.OrgID, cfg.GCPGARImageConfig, cfg.ImageURL)
	if err != nil {
		return fmt.Errorf("authorize GAR source: %w", err)
	}
	source.RegistryType = configs.OCIRegistryTypePrivateOCI
	source.LoginServer = cfg.GCPGARImageConfig.GCPRegion + "-docker.pkg.dev"
	source.Repository = strings.TrimPrefix(cfg.ImageURL, source.LoginServer+"/")
	source.ServiceAccountEmail = ""
	source.WorkloadIdentityProvider = ""
	source.OCIAuth = &configs.OCIRegistryAuth{Username: "oauth2accesstoken", Password: token.AccessToken}
	return nil
}
