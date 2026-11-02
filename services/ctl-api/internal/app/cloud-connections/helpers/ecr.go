package helpers

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/aws/credentials"
	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (h *Helpers) ECRAuthorization(ctx context.Context, orgID string, cfg *app.AWSECRImageConfig, imageURL string) (*ecr.Authorization, error) {
	connection, err := cloudconnections.ECRConnection(ctx, h.db, orgID, cfg, imageURL)
	if err != nil {
		return nil, err
	}
	if connection == nil {
		if !h.cfg.LegacyImageAuthEnabled {
			return nil, cloudconnections.ErrLegacyImageAuthDisabled
		}
		creds := h.legacyECRCredentials(cfg)
		token, err := h.imageAuth.GetECRAccessToken(ctx, &sharedactivities.GetECRAccessTokenRequest{Credentials: creds})
		if err != nil {
			return nil, err
		}
		return &ecr.Authorization{ServerAddress: "https://" + strings.TrimPrefix(token.ServerAddress, "https://"), Username: token.Username, RegistryToken: token.Password}, nil
	}
	creds, err := h.Credentials(ctx, connection, "nuon-ecr-pull")
	if err != nil {
		return nil, err
	}
	creds.Region = cfg.AWSRegion
	client, err := ecr.New(validator.New(), ecr.WithCredentials(creds))
	if err != nil {
		return nil, err
	}
	return client.GetAuthorization(ctx)
}

func (h *Helpers) legacyECRCredentials(cfg *app.AWSECRImageConfig) *credentials.Config {
	role := &credentials.AssumeRoleConfig{
		RoleARN: cfg.IAMRoleARN, SessionName: "container-image-build", SessionDurationSeconds: 30 * 60,
		UseGCPOIDC: h.cfg.IsGCP(),
	}
	if h.cfg.IsAWS() && h.cfg.ManagementIAMRoleARN != "" {
		role.TwoStepConfig = &assumerole.TwoStepConfig{IAMRoleARN: h.cfg.ManagementIAMRoleARN}
	}
	return &credentials.Config{Region: cfg.AWSRegion, AssumeRole: role}
}
