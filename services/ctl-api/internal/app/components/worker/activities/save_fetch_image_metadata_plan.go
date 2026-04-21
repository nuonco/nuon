package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveFetchImageMetadataPlanRequest struct {
	JobID   string `validate:"required"`
	BuildID string `validate:"required"`
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

	srcRepo, err := a.getSourceRepository(extImgCfg)
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

func (a *Activities) getSourceRepository(cfg *app.ExternalImageComponentConfig) (*configs.OCIRegistryRepository, error) {
	if cfg.AWSECRImageConfig != nil {
		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeECR,
			Repository:   cfg.ImageURL,
			Region:       cfg.AWSECRImageConfig.AWSRegion,

			ECRAuth: &credentials.Config{
				Region: cfg.AWSECRImageConfig.AWSRegion,
				AssumeRole: &credentials.AssumeRoleConfig{
					RoleARN:                cfg.AWSECRImageConfig.IAMRoleARN,
					SessionName:            "fetch-image-metadata",
					SessionDurationSeconds: 30 * 60,
					UseGCPOIDC:             a.cfg.IsGCP(),
				},
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
		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeACR,
			Repository:   cfg.ImageURL,
			LoginServer:  cfg.AzureACRImageConfig.RegistryURL,
			ACRAuth: &azurecredentials.Config{
				UseDefault: true,
			},
		}, nil
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   cfg.ImageURL,
		OCIAuth:      &configs.OCIRegistryAuth{},
	}, nil
}
