package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveFetchImageMetadataPlanRequest struct {
	JobID               string `validate:"required"`
	BuildID             string `validate:"required"`
	IsControlPlaneBuild bool
}

// @temporal-gen-v2 activity
// @max-retries 2
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 30s
func (a *Activities) SaveFetchImageMetadataPlan(ctx context.Context, req *SaveFetchImageMetadataPlanRequest) error {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(
		zap.String("job_id", req.JobID),
		zap.String("build_id", req.BuildID),
	)

	l.Info("creating fetch image metadata plan")

	build, err := a.getComponentBuildWithExternalImageConfig(ctx, req.BuildID)
	if err != nil {
		return errors.Wrap(err, "unable to get component build")
	}

	extImgCfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if extImgCfg == nil {
		return fmt.Errorf("build %s does not have external image config", req.BuildID)
	}

	srcRepo, err := a.getSourceRepository(extImgCfg, req.IsControlPlaneBuild, build.ComponentConfigConnection.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get source repository")
	}

	plan := &plantypes.FetchImageMetadataPlan{
		Registry:                    srcRepo,
		Tag:                         extImgCfg.Tag,
		IncludeIndex:                true,
		IncludeAttestationManifests: true,
		IncludeAttestationLayers:    true,
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		return errors.Wrap(err, "unable to marshal plan")
	}

	compositePlan := plantypes.CompositePlan{
		FetchImageMetadataPlan: plan,
	}

	if err := a.runnersHelpers.WriteJobPlan(ctx, req.JobID, planJSON, compositePlan); err != nil {
		return fmt.Errorf("unable to write job plan: %w", err)
	}

	l.Info("fetch image metadata plan saved successfully")
	return nil
}

func (a *Activities) getSourceRepository(cfg *app.ExternalImageComponentConfig, isControlPlaneBuild bool, componentID string) (*configs.OCIRegistryRepository, error) {
	if cfg.AWSECRImageConfig != nil {
		assumeRole := &credentials.AssumeRoleConfig{
			RoleARN:                cfg.AWSECRImageConfig.IAMRoleARN,
			SessionName:            "fetch-image-metadata",
			SessionDurationSeconds: 30 * 60,
			UseGCPOIDC:             a.cfg.IsGCP(),
		}

		// Control-plane metadata jobs run as the ctl-api pod identity, which
		// the vendor's ECR pull role does not trust — vendors grant the Nuon
		// management account, so hop through the management role first (the
		// identity the org runner presented as). Org-runner jobs run in the
		// customer account and assume the pull role directly.
		if isControlPlaneBuild && a.cfg.IsAWS() && a.cfg.ManagementIAMRoleARN != "" {
			assumeRole.TwoStepConfig = &assumerole.TwoStepConfig{
				IAMRoleARN: a.cfg.ManagementIAMRoleARN,
			}
		}

		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeECR,
			Repository:   cfg.ImageURL,
			Region:       cfg.AWSECRImageConfig.AWSRegion,

			ECRAuth: &credentials.Config{
				Region:     cfg.AWSECRImageConfig.AWSRegion,
				AssumeRole: assumeRole,
			},
		}, nil
	}

	if cfg.GCPGARImageConfig != nil {
		garLoginServer := fmt.Sprintf("%s-docker.pkg.dev", cfg.GCPGARImageConfig.GCPRegion)
		return &configs.OCIRegistryRepository{
			RegistryType:             configs.OCIRegistryTypeGAR,
			Repository:               cfg.ImageURL,
			Region:                   cfg.GCPGARImageConfig.GCPRegion,
			LoginServer:              garLoginServer,
			ServiceAccountEmail:      cfg.GCPGARImageConfig.ServiceAccountEmail,
			WorkloadIdentityProvider: cfg.GCPGARImageConfig.WorkloadIdentityProvider,
		}, nil
	}

	if cfg.AzureACRImageConfig != nil {
		acrCfg := &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeACR,
			Repository:   cfg.ImageURL,
			LoginServer:  cfg.AzureACRImageConfig.RegistryURL,
			ACRAuth: &azurecredentials.Config{
				UseDefault: true,
			},
		}

		if acr := cfg.AzureACRImageConfig; acr.ClientID != "" || acr.TenantID != "" ||
			acr.ClientSecretName != "" || acr.ClientCertificateName != "" {
			acrCfg.ACRAppRegistration = &configs.ACRAppRegistration{
				ComponentID:           componentID,
				TenantID:              acr.TenantID,
				ClientID:              acr.ClientID,
				ClientSecretName:      acr.ClientSecretName,
				ClientCertificateName: acr.ClientCertificateName,
			}
		}

		return acrCfg, nil
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   cfg.ImageURL,
		OCIAuth:      &configs.OCIRegistryAuth{},
	}, nil
}
