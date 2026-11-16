package helpers

import (
	"context"
	"strings"

	"golang.org/x/oauth2"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/aws/credentials"
	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func legacyECRCredentialsConfig(cloud *internal.Config, cfg *app.AWSECRImageConfig) *credentials.Config {
	role := &credentials.AssumeRoleConfig{
		RoleARN: cfg.IAMRoleARN, SessionName: "container-image-build", SessionDurationSeconds: 30 * 60,
		UseGCPOIDC: cloud.IsGCP(),
	}
	if cloud.IsAWS() && cloud.ManagementIAMRoleARN != "" {
		role.TwoStepConfig = &assumerole.TwoStepConfig{IAMRoleARN: cloud.ManagementIAMRoleARN}
	}
	return &credentials.Config{Region: cfg.AWSRegion, AssumeRole: role}
}

func (h *Helpers) legacyECRAuthorization(ctx context.Context, cfg *app.AWSECRImageConfig) (*ecr.Authorization, error) {
	token, err := h.imageAuth.GetECRAccessToken(ctx, &sharedactivities.GetECRAccessTokenRequest{
		Credentials: legacyECRCredentialsConfig(h.cfg, cfg),
	})
	if err != nil {
		return nil, err
	}
	return &ecr.Authorization{ServerAddress: "https://" + strings.TrimPrefix(token.ServerAddress, "https://"), Username: token.Username, RegistryToken: token.Password}, nil
}

func (h *Helpers) legacyGARAuthorization(ctx context.Context, cfg *app.GCPGARImageConfig) (*oauth2.Token, error) {
	token, err := h.imageAuth.GetGARAccessToken(ctx, &sharedactivities.GetGARAccessTokenRequest{
		ServiceAccountEmail: cfg.ServiceAccountEmail, WorkloadIdentityProvider: cfg.WorkloadIdentityProvider,
	})
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{AccessToken: token.Password, TokenType: "Bearer"}, nil
}
