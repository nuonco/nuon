package controlplanejob

import (
	"context"
	"fmt"
	"strings"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (a *Activities) authorizeACRSource(ctx context.Context, jobID string, plan *plantypes.CompositePlan) error {
	var source *configs.OCIRegistryRepository
	if plan.BuildPlan != nil && plan.BuildPlan.ContainerImagePullPlan != nil {
		source = plan.BuildPlan.ContainerImagePullPlan.RepoCfg
	}
	if plan.FetchImageMetadataPlan != nil {
		source = plan.FetchImageMetadataPlan.Registry
	}
	if source == nil || source.RegistryType != configs.OCIRegistryTypeACR {
		return nil
	}
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.OrgID == "" || job.OwnerID == "" || job.OwnerType != "component_builds" {
		return fmt.Errorf("ACR source requires an org-owned component build")
	}
	var build app.ComponentBuild
	if err := a.db.WithContext(ctx).
		Where(app.ComponentBuild{ID: job.OwnerID, OrgID: job.OrgID}).
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AzureACRImageConfig").
		First(&build).Error; err != nil {
		return fmt.Errorf("unable to get ACR source build: %w", err)
	}
	cfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if cfg == nil || cfg.AzureACRImageConfig == nil || source.Repository != cfg.ImageURL {
		return fmt.Errorf("component build has no matching ACR source configuration")
	}
	if cfg.AzureACRImageConfig.UsesAmbientCredentials() && !a.cfg.LegacyImageAuthEnabled {
		return cloudconnections.ErrLegacyImageAuthDisabled
	}
	if !cfg.AzureACRImageConfig.UsesCloudConnection() {
		return nil
	}
	auth, err := a.cloudConnectionsHelpers.ACRAuthorization(ctx, build.OrgID, cfg.AzureACRImageConfig, cfg.ImageURL)
	if err != nil {
		return fmt.Errorf("authorize ACR source: %w", err)
	}
	source.RegistryType = configs.OCIRegistryTypePrivateOCI
	source.LoginServer = cfg.AzureACRImageConfig.RegistryURL
	source.Repository = strings.TrimPrefix(cfg.ImageURL, source.LoginServer+"/")
	source.ACRAuth = nil
	source.ACRAppRegistration = nil
	source.OCIAuth = &configs.OCIRegistryAuth{Username: auth.Username, Password: auth.Password}
	return nil
}
