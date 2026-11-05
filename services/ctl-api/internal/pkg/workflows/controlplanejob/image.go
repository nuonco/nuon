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

func imageSource(plan *plantypes.CompositePlan) *configs.OCIRegistryRepository {
	if plan.BuildPlan != nil && plan.BuildPlan.ContainerImagePullPlan != nil {
		return plan.BuildPlan.ContainerImagePullPlan.RepoCfg
	}
	if plan.FetchImageMetadataPlan != nil {
		return plan.FetchImageMetadataPlan.Registry
	}
	return nil
}

func (a *Activities) authorizeImageSource(ctx context.Context, jobID string, plan *plantypes.CompositePlan) error {
	source := imageSource(plan)
	switch {
	case source == nil:
		return nil
	case source.RegistryType != configs.OCIRegistryTypeECR && source.RegistryType != configs.OCIRegistryTypeGAR && source.RegistryType != configs.OCIRegistryTypeACR:
		return nil
	}
	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.OrgID == "" || job.OwnerID == "" || job.OwnerType != "component_builds" {
		return fmt.Errorf("%s source requires an org-owned component build", source.RegistryType)
	}
	var build app.ComponentBuild
	if err := a.db.WithContext(ctx).
		Where(app.ComponentBuild{ID: job.OwnerID, OrgID: job.OrgID}).
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AWSECRImageConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.GCPGARImageConfig").
		Preload("ComponentConfigConnection.ExternalImageComponentConfig.AzureACRImageConfig").
		First(&build).Error; err != nil {
		return fmt.Errorf("unable to get %s source build: %w", source.RegistryType, err)
	}
	cfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if cfg == nil || source.Repository != cfg.ImageURL {
		return fmt.Errorf("component build has no matching %s source configuration", source.RegistryType)
	}
	switch source.RegistryType {
	case configs.OCIRegistryTypeECR:
		if cfg.AWSECRImageConfig == nil {
			return fmt.Errorf("component build has no matching %s source configuration", source.RegistryType)
		}
		return a.authorizeECRSource(ctx, source, build.OrgID, cfg.AWSECRImageConfig, cfg.ImageURL)
	case configs.OCIRegistryTypeGAR:
		if cfg.GCPGARImageConfig == nil {
			return fmt.Errorf("component build has no matching %s source configuration", source.RegistryType)
		}
		return a.authorizeGARSource(ctx, source, build.OrgID, cfg.GCPGARImageConfig, cfg.ImageURL)
	case configs.OCIRegistryTypeACR:
		if cfg.AzureACRImageConfig == nil {
			return fmt.Errorf("component build has no matching %s source configuration", source.RegistryType)
		}
		return a.authorizeACRSource(ctx, source, build.OrgID, cfg.AzureACRImageConfig, cfg.ImageURL)
	}
	return nil
}

func (a *Activities) authorizeECRSource(ctx context.Context, source *configs.OCIRegistryRepository, orgID string, cfg *app.AWSECRImageConfig, imageURL string) error {
	auth, err := a.cloudConnectionsHelpers.ECRAuthorization(ctx, orgID, cfg, imageURL)
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

func (a *Activities) authorizeGARSource(ctx context.Context, source *configs.OCIRegistryRepository, orgID string, cfg *app.GCPGARImageConfig, imageURL string) error {
	token, err := a.cloudConnectionsHelpers.GARAuthorization(ctx, orgID, cfg, imageURL)
	if err != nil {
		return fmt.Errorf("authorize GAR source: %w", err)
	}
	source.RegistryType = configs.OCIRegistryTypePrivateOCI
	source.LoginServer = cfg.GCPRegion + "-docker.pkg.dev"
	source.Repository = strings.TrimPrefix(imageURL, source.LoginServer+"/")
	source.ServiceAccountEmail = ""
	source.WorkloadIdentityProvider = ""
	source.OCIAuth = &configs.OCIRegistryAuth{Username: "oauth2accesstoken", Password: token.AccessToken}
	return nil
}

func (a *Activities) authorizeACRSource(ctx context.Context, source *configs.OCIRegistryRepository, orgID string, cfg *app.AzureACRImageConfig, imageURL string) error {
	auth, err := a.cloudConnectionsHelpers.ACRAuthorization(ctx, orgID, cfg, imageURL)
	if err != nil {
		return fmt.Errorf("authorize ACR source: %w", err)
	}
	if auth == nil {
		return nil
	}
	source.RegistryType = configs.OCIRegistryTypePrivateOCI
	_, repository, _ := strings.Cut(imageURL, "/")
	source.LoginServer = strings.ToLower(cfg.RegistryURL)
	source.Repository = repository
	source.ACRAuth = nil
	source.ACRAppRegistration = nil
	source.OCIAuth = &configs.OCIRegistryAuth{Username: auth.Username, Password: auth.Password}
	return nil
}
