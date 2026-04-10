package activities

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/pkg/oci/metadata"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	"golang.org/x/oauth2/google"
)

type FetchImageMetadataRequest struct {
	BuildID string `validate:"required"`
}

type FetchImageMetadataResult struct {
	Metadata *metadata.ImageMetadata `json:"metadata" temporaljson:"metadata,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 2
// @schedule-to-close-timeout 5m
// @start-to-close-timeout 4m
func (a *Activities) FetchImageMetadata(ctx context.Context, req *FetchImageMetadataRequest) (*FetchImageMetadataResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("build_id", req.BuildID))

	l.Info("fetching image metadata for policy evaluation")

	build, err := a.getComponentBuildWithExternalImageConfig(ctx, req.BuildID)
	if err != nil {
		l.Error("unable to get component build", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get component build")
	}

	extImgCfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if extImgCfg == nil {
		return nil, fmt.Errorf("build %s does not have external image config", req.BuildID)
	}

	l = l.With(
		zap.String("image_url", extImgCfg.ImageURL),
		zap.String("tag", extImgCfg.Tag),
	)

	fetchOpts := &metadata.FetchOptions{
		Image:                       extImgCfg.ImageURL,
		Tag:                         extImgCfg.Tag,
		IncludeIndex:                true,
		IncludeAttestationManifests: true,
		IncludeAttestationLayers:    true,
	}

	if extImgCfg.AWSECRImageConfig != nil {
		l.Debug("fetching ECR credentials for private registry")
		auth, err := a.getECRAuth(ctx, extImgCfg.AWSECRImageConfig)
		if err != nil {
			l.Error("unable to get ECR authorization", zap.Error(err))
			return nil, errors.Wrap(err, "unable to get ECR authorization")
		}
		fetchOpts.Auth = auth
	}

	if extImgCfg.GCPGARImageConfig != nil {
		l.Debug("fetching GAR credentials for private registry")
		auth, err := a.getGARAuth(ctx, extImgCfg.GCPGARImageConfig)
		if err != nil {
			l.Error("unable to get GAR authorization", zap.Error(err))
			return nil, errors.Wrap(err, "unable to get GAR authorization")
		}
		fetchOpts.Auth = auth
	}

	if extImgCfg.AzureACRImageConfig != nil {
		l.Debug("fetching ACR credentials for private registry")
		auth, err := a.getACRAuth(ctx, extImgCfg.AzureACRImageConfig)
		if err != nil {
			l.Error("unable to get ACR authorization", zap.Error(err))
			return nil, errors.Wrap(err, "unable to get ACR authorization")
		}
		fetchOpts.Auth = auth
	}

	l.Info("fetching image referrers for metadata")
	imgMetadata, err := metadata.FetchImageMetadata(ctx, fetchOpts)
	if err != nil {
		l.Error("unable to fetch image metadata", zap.Error(err))
		return nil, errors.Wrap(err, "unable to fetch image metadata")
	}

	l.Info("image metadata fetched successfully",
		zap.String("digest", imgMetadata.Digest),
		zap.Bool("signed", imgMetadata.Signed),
		zap.Bool("has_sbom", imgMetadata.SBOM != nil && imgMetadata.SBOM.Present),
		zap.Int("attestations_count", len(imgMetadata.Attestations)),
	)

	return &FetchImageMetadataResult{
		Metadata: imgMetadata,
	}, nil
}

func (a *Activities) getComponentBuildWithExternalImageConfig(ctx context.Context, buildID string) (*app.ComponentBuild, error) {
	var bld app.ComponentBuild

	res := a.db.WithContext(ctx).
		Scopes(helpers.PreloadComponentBuildConfig).
		First(&bld, "id = ?", buildID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get component build")
	}

	return &bld, nil
}

func (a *Activities) getACRAuth(ctx context.Context, acrCfg *app.AzureACRImageConfig) (*metadata.RegistryAuth, error) {
	token, err := acr.GetRepositoryToken(ctx, acrCfg.CredentialsConfig(), acrCfg.RegistryURL, zap.L())
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ACR repository token")
	}

	return &metadata.RegistryAuth{
		ServerAddress: "https://" + acrCfg.RegistryURL,
		Username:      acr.DefaultACRUsername,
		Password:      token,
	}, nil
}

func (a *Activities) getGARAuth(ctx context.Context, garCfg *app.GCPGARImageConfig) (*metadata.RegistryAuth, error) {
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get GCP token source for GAR")
	}

	token, err := ts.Token()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get GCP access token for GAR")
	}

	host := garCfg.GCPRegion + "-docker.pkg.dev"
	return &metadata.RegistryAuth{
		ServerAddress: "https://" + host,
		Username:      "oauth2accesstoken",
		Password:      token.AccessToken,
	}, nil
}

func (a *Activities) getECRAuth(ctx context.Context, ecrCfg *app.AWSECRImageConfig) (*metadata.RegistryAuth, error) {
	v := validator.New()

	credsCfg := &credentials.Config{
		Region: ecrCfg.AWSRegion,
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     ecrCfg.IAMRoleARN,
			SessionName: "ctl-api-image-metadata-fetch",
			UseGCPOIDC:  a.cfg.IsGCP(),
		},
	}

	ecrClient, err := ecr.New(v,
		ecr.WithCredentials(credsCfg),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ECR client")
	}

	auth, err := ecrClient.GetAuthorization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ECR authorization")
	}

	return &metadata.RegistryAuth{
		ServerAddress: auth.ServerAddress,
		Username:      auth.Username,
		Password:      auth.RegistryToken,
	}, nil
}
