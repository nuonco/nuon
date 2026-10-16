package registryauth

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type refreshSource struct {
	key  string
	mint func(ctx context.Context, l *zap.Logger) (*configs.OCIRegistryAuth, error)
}

func (i *Issuer) refreshSource(orgID string, cfg *configs.OCIRegistryRepository) *refreshSource {
	if cfg == nil || cfg.OCIAuth == nil || cfg.OCIAuth.Password == "" {
		return nil
	}

	switch cfg.RegistryType {
	case configs.OCIRegistryTypeGAR:
		sa, wip := cfg.ServiceAccountEmail, cfg.WorkloadIdentityProvider
		return &refreshSource{
			key: strings.Join([]string{"gar", sa, wip}, "|"),
			mint: func(ctx context.Context, l *zap.Logger) (*configs.OCIRegistryAuth, error) {
				return toAuth(i.GAR(ctx, l, sa, wip))
			},
		}

	case configs.OCIRegistryTypeACR:
		if cfg.ACRAppRegistration == nil {
			return nil
		}
		reg := cfg.ACRAppRegistration
		req := ACRRequest{
			ComponentID:           reg.ComponentID,
			LoginServer:           cfg.LoginServer,
			TenantID:              reg.TenantID,
			ClientID:              reg.ClientID,
			ClientSecretName:      reg.ClientSecretName,
			ClientCertificateName: reg.ClientCertificateName,
		}
		return &refreshSource{
			key: strings.Join([]string{
				"acr", req.LoginServer, req.ComponentID, req.TenantID, req.ClientID,
				req.ClientSecretName, req.ClientCertificateName,
			}, "|"),
			mint: func(ctx context.Context, l *zap.Logger) (*configs.OCIRegistryAuth, error) {
				return toAuth(i.ACR(ctx, l, req))
			},
		}

	case configs.OCIRegistryTypePrivateOCI:
		region, ok := i.orgRegistryRegion(cfg.LoginServer)
		if !ok {
			return nil
		}
		return &refreshSource{
			key: strings.Join([]string{"org", orgID, region}, "|"),
			mint: func(ctx context.Context, l *zap.Logger) (*configs.OCIRegistryAuth, error) {
				tok, err := i.OrgRegistry(ctx, l, orgID, region)
				if err != nil {
					return nil, err
				}
				return &configs.OCIRegistryAuth{Username: tok.Username, Password: tok.RegistryToken}, nil
			},
		}
	}

	return nil
}

func (i *Issuer) orgRegistryRegion(loginServer string) (string, bool) {
	host := strings.TrimPrefix(strings.TrimSpace(loginServer), "https://")
	if host == "" {
		return "", false
	}

	switch {
	case i.cfg.IsGCP():
		return "", host == i.managementGARHost()
	case i.cfg.IsAzure():
		return "", host == strings.TrimPrefix(i.cfg.ManagementACRRegistryURL, "https://")
	}

	if i.cfg.ManagementECRRegistryID == "" {
		return "", false
	}
	rest, ok := strings.CutPrefix(host, i.cfg.ManagementECRRegistryID+".dkr.ecr.")
	if !ok {
		return "", false
	}
	region, _, _ := strings.Cut(rest, ".")
	return region, region != ""
}

func toAuth(tok *Token, err error) (*configs.OCIRegistryAuth, error) {
	if err != nil {
		return nil, err
	}
	return &configs.OCIRegistryAuth{Username: tok.Username, Password: tok.Password}, nil
}
