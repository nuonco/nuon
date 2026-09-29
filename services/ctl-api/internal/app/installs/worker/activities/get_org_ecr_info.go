package activities

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2/google"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	ecrauthorization "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/azure/acr"
)

type OrgECRAccessInfo struct {
	RegistryID    string
	Username      string
	RegistryToken string
	ServerAddress string
	Region        string
}

// @temporal-gen-v2 activity
func (a *Activities) GetOrgECRAccessInfo(ctx context.Context, orgID string) (*OrgECRAccessInfo, error) {
	if a.cfg.IsGCP() {
		return a.getOrgGARAccessInfo(ctx)
	}
	if a.cfg.IsAzure() {
		return a.getOrgACRAccessInfo(ctx)
	}

	ecr, err := ecrauthorization.New(a.v,
		ecrauthorization.WithCredentials(&credentials.Config{
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:     a.cfg.ManagementIAMRoleARN,
				SessionName: fmt.Sprintf("oci-sync-%s", orgID),
			},
		}),
		ecrauthorization.WithRegistryID(a.cfg.ManagementECRRegistryID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create ecrauthorizer for image sync: %w", err)
	}

	ecrAuth, err := ecr.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &OrgECRAccessInfo{
		RegistryID:    a.cfg.ManagementECRRegistryID,
		Username:      ecrAuth.Username,
		RegistryToken: ecrAuth.RegistryToken,
		ServerAddress: ecrAuth.ServerAddress,
	}, nil
}

func (a *Activities) getOrgGARAccessInfo(ctx context.Context) (*OrgECRAccessInfo, error) {
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP token source for GAR: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP access token for GAR: %w", err)
	}

	repoURL := a.cfg.ManagementGARRepositoryURL
	host := repoURL
	if idx := strings.Index(repoURL, "/"); idx != -1 {
		host = repoURL[:idx]
	}

	return &OrgECRAccessInfo{
		RegistryID:    repoURL,
		Username:      "oauth2accesstoken",
		RegistryToken: token.AccessToken,
		ServerAddress: "https://" + host,
	}, nil
}

func (a *Activities) getOrgACRAccessInfo(ctx context.Context) (*OrgECRAccessInfo, error) {
	acrService := a.cfg.ManagementACRRegistryURL

	token, err := acr.GetRepositoryToken(ctx, nil, acrService, zap.L())
	if err != nil {
		return nil, fmt.Errorf("unable to get ACR repository token: %w", err)
	}

	return &OrgECRAccessInfo{
		RegistryID:    acrService,
		Username:      acr.DefaultACRUsername,
		RegistryToken: token,
		ServerAddress: "https://" + acrService,
	}, nil
}
