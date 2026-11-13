package build

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var eksAccessPolicyARN = regexp.MustCompile(`^arn:aws[a-z-]*:eks::aws:cluster-access-policy/[A-Za-z0-9]+$`)

func EKSAccess(access *config.AppAWSIAMRoleEKSAccess) *app.AppAWSIAMRoleEKSAccess {
	if access == nil {
		return nil
	}

	policies := make([]app.AppAWSIAMRoleEKSAccessPolicy, 0, len(access.Policies))
	for _, policy := range access.Policies {
		scope := policy.AccessScope
		if scope == "" {
			scope = config.EKSAccessScopeCluster
		}
		policies = append(policies, app.AppAWSIAMRoleEKSAccessPolicy{
			PolicyARN:   policy.PolicyARN,
			AccessScope: scope,
			Namespaces:  policy.Namespaces,
		})
	}

	return &app.AppAWSIAMRoleEKSAccess{
		ClusterName:      access.ClusterName,
		KubernetesGroups: access.KubernetesGroups,
		Policies:         policies,
	}
}

func ValidateEKSAccess(roles []app.AppAWSIAMRoleConfig, stackType string) error {
	for _, role := range roles {
		access := role.EKSAccess
		if access == nil {
			continue
		}

		if stackType != awsCloudFormationStackType {
			return fmt.Errorf("role %q: eks_access requires stack type %q, got %q", role.Name, awsCloudFormationStackType, stackType)
		}
		if err := validateEKSAccessShape(role); err != nil {
			return err
		}
	}
	return nil
}

func validateEKSAccessShape(role app.AppAWSIAMRoleConfig) error {
	access := role.EKSAccess
	if access == nil {
		return nil
	}

	platform := strings.ToLower(role.CloudPlatform)
	if platform != "" && platform != "aws" {
		return fmt.Errorf("role %q: eks_access is AWS-only (cloud_platform %q)", role.Name, role.CloudPlatform)
	}
	if role.Type != app.AWSIAMRoleTypeBreakGlass && role.Type != app.AWSIAMRoleTypeCustom {
		return fmt.Errorf("role %q: eks_access is only supported on break glass and custom roles; declare %s access entries in the sandbox", role.Name, role.Type)
	}
	if strings.TrimSpace(access.ClusterName) == "" {
		return fmt.Errorf("role %q: eks_access.cluster_name is required", role.Name)
	}
	if err := render.ValidateTextTemplate(access.ClusterName); err != nil {
		return fmt.Errorf("role %q: eks_access.cluster_name is not a valid template: %w", role.Name, err)
	}
	if len(access.Policies) == 0 && len(access.KubernetesGroups) == 0 {
		return fmt.Errorf("role %q: eks_access grants nothing; set at least one policy or kubernetes group", role.Name)
	}

	seen := make(map[string]struct{}, len(access.Policies))
	for _, policy := range access.Policies {
		if !eksAccessPolicyARN.MatchString(policy.PolicyARN) {
			return fmt.Errorf("role %q: eks_access policy_arn %q is not an EKS cluster access policy ARN", role.Name, policy.PolicyARN)
		}
		if _, ok := seen[policy.PolicyARN]; ok {
			return fmt.Errorf("role %q: eks_access policy %q is listed more than once", role.Name, policy.PolicyARN)
		}
		seen[policy.PolicyARN] = struct{}{}

		switch policy.AccessScope {
		case "", config.EKSAccessScopeCluster:
			if len(policy.Namespaces) > 0 {
				return fmt.Errorf("role %q: eks_access policy %q has namespaces but access_scope is cluster", role.Name, policy.PolicyARN)
			}
		case config.EKSAccessScopeNamespace:
			if len(policy.Namespaces) == 0 {
				return fmt.Errorf("role %q: eks_access policy %q has access_scope namespace but no namespaces", role.Name, policy.PolicyARN)
			}
		default:
			return fmt.Errorf("role %q: eks_access policy %q has unknown access_scope %q; use %q or %q", role.Name, policy.PolicyARN, policy.AccessScope, config.EKSAccessScopeCluster, config.EKSAccessScopeNamespace)
		}
	}
	return nil
}
