package controlplanejob

import (
	"context"
	"fmt"
	"strings"

	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (a *Activities) authorizeECRSource(ctx context.Context, jobID string, plan *plantypes.CompositePlan) error {
	var source *configs.OCIRegistryRepository
	if plan.BuildPlan != nil && plan.BuildPlan.ContainerImagePullPlan != nil {
		source = plan.BuildPlan.ContainerImagePullPlan.RepoCfg
	}
	if plan.FetchImageMetadataPlan != nil {
		source = plan.FetchImageMetadataPlan.Registry
	}
	if source == nil || source.RegistryType != configs.OCIRegistryTypeECR {
		return nil
	}
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.OrgID == "" || job.OwnerID == "" || job.OwnerType != "component_builds" {
		return fmt.Errorf("ECR source requires an org-owned component build")
	}
	var build app.ComponentBuild
	if err := a.db.WithContext(ctx).
		Where(app.ComponentBuild{ID: job.OwnerID, OrgID: job.OrgID}).
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AWSECRImageConfig").
		First(&build).Error; err != nil {
		return fmt.Errorf("unable to get ECR source build: %w", err)
	}
	cfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if cfg == nil || cfg.AWSECRImageConfig == nil || source.Repository != cfg.ImageURL {
		return fmt.Errorf("component build has no matching ECR source configuration")
	}
	auth, err := a.cloudConnectionsHelpers.ECRAuthorization(ctx, build.OrgID, cfg.AWSECRImageConfig, cfg.ImageURL)
	if err != nil {
		return fmt.Errorf("authorize ECR source: %w", err)
	}
	repo, err := ecr.TrimRepositoryName(source.Repository, auth.ServerAddress)
	if err != nil {
		return fmt.Errorf("invalid ECR source repository: %w", err)
	}
	source.RegistryType = configs.OCIRegistryTypePrivateOCI
	source.Repository = repo
	source.LoginServer = strings.TrimPrefix(auth.ServerAddress, "https://")
	source.ECRAuth = nil
	source.OCIAuth = &configs.OCIRegistryAuth{Username: auth.Username, Password: auth.RegistryToken}
	return nil
}
