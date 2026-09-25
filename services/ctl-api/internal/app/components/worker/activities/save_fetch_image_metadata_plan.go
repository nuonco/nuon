package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

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

	srcRepo, err := a.getSourceRepository(ctx, extImgCfg, build.ComponentConfigConnection.ComponentID)
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

func (a *Activities) getSourceRepository(ctx context.Context, cfg *app.ExternalImageComponentConfig, componentID string) (*configs.OCIRegistryRepository, error) {
	if cfg.AWSECRImageConfig != nil {
		auth, err := a.cloudConnections.ECRCredentials(ctx, cfg.AWSECRImageConfig.CloudConnection, "fetch-image-metadata")
		if err != nil {
			return nil, err
		}

		return &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypeECR,
			Repository:   cfg.ImageURL,
			Region:       cfg.AWSECRImageConfig.AWSRegion,

			ECRAuth: auth,
		}, nil
	}

	if cfg.GCPGARImageConfig != nil {
		garLoginServer := fmt.Sprintf("%s-docker.pkg.dev", cfg.GCPGARImageConfig.GCPRegion)
		repository := &configs.OCIRegistryRepository{
			RegistryType:             configs.OCIRegistryTypeGAR,
			Repository:               cfg.ImageURL,
			Region:                   cfg.GCPGARImageConfig.GCPRegion,
			LoginServer:              garLoginServer,
			ServiceAccountEmail:      cfg.GCPGARImageConfig.ServiceAccountEmail,
			WorkloadIdentityProvider: cfg.GCPGARImageConfig.WorkloadIdentityProvider,
		}
		if connection := cfg.GCPGARImageConfig.CloudConnection; connection != nil && connection.AuthMode != app.CloudConnectionAuthModeLegacy {
			token, err := a.cloudConnections.GCPAccessToken(ctx, connection)
			if err != nil {
				return nil, fmt.Errorf("get GCP cloud connection credentials: %w", err)
			}
			repository.OCIAuth = &configs.OCIRegistryAuth{Username: "oauth2accesstoken", Password: token.AccessToken}
		}
		return repository, nil
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

		if acr := cfg.AzureACRImageConfig; acr.CloudConnectionID != "" || acr.ClientID != "" || acr.TenantID != "" ||
			acr.ClientSecretName != "" || acr.ClientCertificateName != "" {
			acrCfg.ACRAppRegistration = &configs.ACRAppRegistration{
				ComponentID:           componentID,
				ConnectionID:          acr.CloudConnectionID,
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
