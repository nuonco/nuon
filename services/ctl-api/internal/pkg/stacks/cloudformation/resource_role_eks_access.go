package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/eks"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func eksAccessClusterParamName(role app.AppAWSIAMRoleConfig) string {
	return role.CloudFormationStackName + "EKSClusterName"
}

func eksAccessConditionName(role app.AppAWSIAMRoleConfig) string {
	return role.CloudFormationStackName + "EKSAccess"
}

func eksAccessEntryLogicalID(role app.AppAWSIAMRoleConfig) string {
	return role.CloudFormationStackName + "EKSAccessEntry"
}

func eksAccessRoles(inp *stacks.TemplateInput) []app.AppAWSIAMRoleConfig {
	seen := map[string]struct{}{}
	out := make([]app.AppAWSIAMRoleConfig, 0)
	groups := [][]app.AppAWSIAMRoleConfig{
		inp.AppCfg.PermissionsConfig.Roles,
		inp.AppCfg.BreakGlassConfig.Roles,
		inp.AppCfg.PermissionsConfig.CustomRoles,
	}
	for _, roles := range groups {
		for _, role := range roles {
			if role.EKSAccess == nil {
				continue
			}
			if _, ok := seen[role.CloudFormationStackName]; ok {
				continue
			}
			seen[role.CloudFormationStackName] = struct{}{}
			out = append(out, role)
		}
	}
	return out
}

func renderEKSClusterName(clusterName string, inp *stacks.TemplateInput) string {
	if inp.InstallState == nil {
		return ""
	}
	stateMap, err := inp.InstallState.AsMap()
	if err != nil {
		return ""
	}
	rendered, err := render.RenderTextV2(clusterName, stateMap)
	if err != nil {
		return ""
	}
	return rendered
}

func (a *Templates) getEKSAccessParameters(inp *stacks.TemplateInput) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter, 0)
	for _, role := range eksAccessRoles(inp) {
		params[eksAccessClusterParamName(role)] = cloudformation.Parameter{
			Type:        "String",
			Default:     renderEKSClusterName(role.EKSAccess.ClusterName, inp),
			Description: ptr(fmt.Sprintf("EKS cluster %s gets access to while it is enabled. Leave empty to skip the access entry.", role.Name)),
		}
	}
	return params
}

func (a *Templates) getEKSAccessConditions(inp *stacks.TemplateInput) map[string]any {
	conditions := make(map[string]any, 0)
	for _, role := range eksAccessRoles(inp) {
		conditions[eksAccessConditionName(role)] = cloudformation.And([]string{
			cloudformation.Equals(cloudformation.Ref(a.roleConditionName(role)), "true"),
			cloudformation.Not([]string{
				cloudformation.Equals(cloudformation.Ref(eksAccessClusterParamName(role)), ""),
			}),
		})
	}
	return conditions
}

func (a *Templates) getEKSAccessParamLabels(inp *stacks.TemplateInput) map[string]any {
	labels := make(map[string]any, 0)
	for _, role := range eksAccessRoles(inp) {
		name := role.DisplayName
		if name == "" {
			name = role.Name
		}
		labels[eksAccessClusterParamName(role)] = name + " EKS cluster"
	}
	return labels
}

func (a *Templates) getEKSAccessResources(inp *stacks.TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	rsrcs := make(map[string]cloudformation.Resource, 0)
	for _, role := range eksAccessRoles(inp) {
		policies := make([]eks.AccessEntry_AccessPolicy, 0, len(role.EKSAccess.Policies))
		for _, policy := range role.EKSAccess.Policies {
			policies = append(policies, eks.AccessEntry_AccessPolicy{
				PolicyArn: policy.PolicyARN,
				AccessScope: &eks.AccessEntry_AccessScope{
					Type:       policy.AccessScope,
					Namespaces: policy.Namespaces,
				},
			})
		}

		rsrcs[eksAccessEntryLogicalID(role)] = &eks.AccessEntry{
			AWSCloudFormationCondition: eksAccessConditionName(role),
			ClusterName:                cloudformation.Ref(eksAccessClusterParamName(role)),
			PrincipalArn:               cloudformation.GetAtt(role.CloudFormationStackName, "Arn"),
			KubernetesGroups:           role.EKSAccess.KubernetesGroups,
			AccessPolicies:             policies,
			Tags:                       t.apply(nil, fmt.Sprintf("%s-eks-access", role.Type)),
		}
	}
	return rsrcs
}
