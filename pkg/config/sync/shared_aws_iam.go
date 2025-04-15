package sync

import (
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) awsIAMPolicyToRequest(policy config.AppAWSIAMPolicy) *models.ServiceAppAWSIAMPolicyConfig {
	return &models.ServiceAppAWSIAMPolicyConfig{
		Contents:          policy.Contents,
		ManagedPolicyName: policy.ManagedPolicyName,
		Name:              policy.Name,
	}
}

func (s sync) awsIAMRoleToRequest(role *config.AppAWSIAMRole) *models.ServiceAppAWSIAMRoleConfig {
	policies := make([]*models.ServiceAppAWSIAMPolicyConfig, 0)
	for _, policy := range role.Policies {
		policies = append(policies, s.awsIAMPolicyToRequest(policy))
	}

	req := &models.ServiceAppAWSIAMRoleConfig{
		Description: generics.ToPtr(role.Description),
		DisplayName: generics.ToPtr(role.DisplayName),
		Name:        generics.ToPtr(role.Name),
		Policies:    policies,
	}

	if role.PermissionsBoundary != "" {
		req.PermissionsBoundary = role.PermissionsBoundary
	}

	return req
}
