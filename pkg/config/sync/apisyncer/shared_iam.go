package apisyncer

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func (s *syncer) iamPolicyToRequest(policy config.AppAWSIAMPolicy) *models.ServiceAppAWSIAMPolicyConfig {
	return &models.ServiceAppAWSIAMPolicyConfig{
		Contents:          policy.Contents,
		GcpPermissions:    policy.GCPPermissions,
		GcpPredefinedRole: policy.GCPPredefinedRole,
		ManagedPolicyName: policy.ManagedPolicyName,
		Name:              policy.Name,
	}
}

func (s *syncer) iamRoleToRequest(role *config.AppAWSIAMRole) *models.ServiceAppAWSIAMRoleConfig {
	policies := make([]*models.ServiceAppAWSIAMPolicyConfig, 0)
	for _, policy := range role.Policies {
		policies = append(policies, s.iamPolicyToRequest(policy))
	}

	req := &models.ServiceAppAWSIAMRoleConfig{
		CloudPlatform: role.CloudPlatform,
		Description:   generics.ToPtr(role.Description),
		DisplayName:   generics.ToPtr(role.DisplayName),
		Name:          generics.ToPtr(role.Name),
		Policies:      policies,
	}

	if role.PermissionsBoundary != "" {
		req.PermissionsBoundary = role.PermissionsBoundary
	}

	return req
}
