package registryauth

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

type OrgRegistryToken struct {
	RegistryID    string
	Username      string
	RegistryToken string
	ServerAddress string
}

func (i *Issuer) OrgRegistry(ctx context.Context, l *zap.Logger, orgID, region string) (*OrgRegistryToken, error) {
	if i.cfg.IsGCP() {
		return i.orgGAR(ctx)
	}
	if i.cfg.IsAzure() {
		return i.orgACR(ctx, l)
	}

	ecr, err := ecrauthorization.New(i.v,
		ecrauthorization.WithCredentials(&credentials.Config{
			Region: region,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:     i.cfg.ManagementIAMRoleARN,
				SessionName: fmt.Sprintf("oci-sync-%s", orgID),
			},
		}),
		ecrauthorization.WithRegistryID(i.cfg.ManagementECRRegistryID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create ecrauthorizer for image sync: %w", err)
	}

	ecrAuth, err := ecr.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &OrgRegistryToken{
		RegistryID:    i.cfg.ManagementECRRegistryID,
		Username:      ecrAuth.Username,
		RegistryToken: ecrAuth.RegistryToken,
		ServerAddress: ecrAuth.ServerAddress,
	}, nil
}

func (i *Issuer) orgGAR(ctx context.Context) (*OrgRegistryToken, error) {
	ts, err := google.DefaultTokenSource(ctx, garScopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP token source for GAR: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP access token for GAR: %w", err)
	}

	return &OrgRegistryToken{
		RegistryID:    i.cfg.ManagementGARRepositoryURL,
		Username:      garUsername,
		RegistryToken: token.AccessToken,
		ServerAddress: "https://" + i.managementGARHost(),
	}, nil
}

func (i *Issuer) orgACR(ctx context.Context, l *zap.Logger) (*OrgRegistryToken, error) {
	acrService := i.cfg.ManagementACRRegistryURL

	token, err := acr.GetRepositoryToken(ctx, nil, acrService, l)
	if err != nil {
		return nil, fmt.Errorf("unable to get ACR repository token: %w", err)
	}

	return &OrgRegistryToken{
		RegistryID:    acrService,
		Username:      acr.DefaultACRUsername,
		RegistryToken: token,
		ServerAddress: "https://" + acrService,
	}, nil
}

func (i *Issuer) managementGARHost() string {
	host, _, _ := strings.Cut(i.cfg.ManagementGARRepositoryURL, "/")
	return host
}
