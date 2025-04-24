package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
)

type OrgECRAccessInfo struct {
	RegistryID    string
	Username      string
	RegistryToken string
	ServerAddress string
	Region        string
}

// @temporal-gen activity
func (a *Activities) GetOrgECRAccessInfo(ctx context.Context, orgID string) (*OrgECRAccessInfo, error) {
	ecr, err := ecrauthorization.New(a.v,
		ecrauthorization.WithCredentials(&credentials.Config{
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:     fmt.Sprintf(a.legacyTFCloudOutputs.OrgsIAMRoleNameTemplateOutputs.InstancesAccess, orgID),
				SessionName: fmt.Sprintf("oci-sync-%s", orgID),
			},
		}),
		ecrauthorization.WithRegistryID(a.legacyTFCloudOutputs.ECR.RegistryID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create ecrauthorizer for image sync: %w", err)
	}

	ecrAuth, err := ecr.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &OrgECRAccessInfo{
		RegistryID:    a.legacyTFCloudOutputs.ECR.RegistryID,
		Username:      ecrAuth.Username,
		RegistryToken: ecrAuth.RegistryToken,
		ServerAddress: ecrAuth.ServerAddress,
	}, nil
}
