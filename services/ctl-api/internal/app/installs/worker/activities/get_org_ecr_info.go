package activities

import (
	"context"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
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
	token, err := a.registryAuth.OrgRegistry(ctx, temporalzap.GetActivityLogger(ctx), orgID, "")
	if err != nil {
		return nil, err
	}

	return &OrgECRAccessInfo{
		RegistryID:    token.RegistryID,
		Username:      token.Username,
		RegistryToken: token.RegistryToken,
		ServerAddress: token.ServerAddress,
	}, nil
}
