package activities

import (
	"context"
	"fmt"
)

type OrgECRConfig struct {
	ManagementIAMRoleARN string
	RegistryID           string
	ServerAddress        string
}

// @temporal-gen-v2 activity
func (a *Activities) GetOrgECRConfig(ctx context.Context, orgID string) (*OrgECRConfig, error) {
	if a.cfg.ManagementIAMRoleARN == "" {
		return nil, fmt.Errorf("management IAM role ARN is not configured")
	}

	return &OrgECRConfig{
		ManagementIAMRoleARN: a.cfg.ManagementIAMRoleARN,
		RegistryID:           a.cfg.ManagementECRRegistryID,
		ServerAddress:        fmt.Sprintf("%s.dkr.ecr.us-west-2.amazonaws.com", a.cfg.ManagementECRRegistryID),
	}, nil
}
