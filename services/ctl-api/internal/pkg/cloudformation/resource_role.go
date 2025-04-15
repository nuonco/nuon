package cloudformation

import (
	"encoding/json"
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Templates) roleConditionName(role app.AppAWSIAMRoleConfig) string {
	return role.CloudFormationStackParamName
}

func (t *Templates) getRoleConditions(inp *TemplateInput) map[string]any {
	conditions := make(map[string]any, 0)

	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		conditions[role.CloudFormationStackParamName] = cloudformation.Equals(cloudformation.Ref(t.roleConditionName(role)), "true")
	}

	return conditions
}

func (a *Templates) getRolesParamLabels(inp *TemplateInput) map[string]any {
	paramLabels := make(map[string]any, 0)
	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		paramLabels[role.CloudFormationStackParamName] = role.DisplayName
	}

	return paramLabels
}

func (a *Templates) getRolesResources(inp *TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	rsrcs := make(map[string]cloudformation.Resource, 0)

	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		managedPolicyARNs := make([]string, 0)
		for _, policy := range role.Policies {
			if policy.ManagedPolicyName == "" {
				continue
			}

			managedPolicyARNs = append(managedPolicyARNs, fmt.Sprintf("arn:aws:iam::aws:policy/%s", policy.ManagedPolicyName))
		}

		rsrcs[role.CloudFormationStackName] = &iam.Role{
			AWSCloudFormationCondition: a.roleConditionName(role),
			RoleName:                   generics.ToPtr(role.Name),
			ManagedPolicyArns:          managedPolicyARNs,
			AssumeRolePolicyDocument: map[string]any{
				"Statement": []map[string]any{
					{
						"Effect": "Allow",
						"Principal": map[string]any{
							"AWS": cloudformation.GetAttPtr("RunnerInstanceRole", "Arn"),
						},
						"Action": "sts:AssumeRole",
					},
				},
			},
			Tags: t.apply(nil, fmt.Sprintf("%s-role", role.Type)),
		}

		// create each policy
		for _, policy := range role.Policies {
			if policy.ManagedPolicyName != "" {
				continue
			}

			rsrcs[policy.CloudFormationStackName] = a.getRolePolicy(role, policy)
		}
	}

	return rsrcs
}

func (a *Templates) getRolePolicy(role app.AppAWSIAMRoleConfig, policy app.AppAWSIAMPolicyConfig) cloudformation.Resource {
	return &iam.Policy{
		AWSCloudFormationCondition: a.roleConditionName(role),
		PolicyName: cloudformation.SubVars(
			fmt.Sprintf(policy.Name),
			map[string]any{"RoleName": cloudformation.Ref(role.CloudFormationStackName)}),
		PolicyDocument: json.RawMessage([]byte(policy.Contents)),
		Roles:          []string{cloudformation.Ref(role.CloudFormationStackName)},
	}
}

func (a *Templates) getRolesParameters(inp *TemplateInput) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter, 0)

	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		params[role.CloudFormationStackParamName] = cloudformation.Parameter{
			Type:    "String",
			Default: true,
			AllowedValues: []any{
				"true",
				"false",
			},
			Description: generics.ToPtr(role.Description),
		}
	}

	return params
}
