package cloudformation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"

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

	for _, role := range inp.AppCfg.BreakGlassConfig.Roles {
		conditions[role.CloudFormationStackParamName] = cloudformation.Equals(cloudformation.Ref(t.roleConditionName(role)), "true")
	}

	return conditions
}

func (a *Templates) getRolesParamLabels(inp *TemplateInput) map[string]any {
	paramLabels := make(map[string]any, 0)
	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		paramLabels[role.CloudFormationStackParamName] = role.DisplayName
	}
	for _, role := range inp.AppCfg.BreakGlassConfig.Roles {
		paramLabels[role.CloudFormationStackParamName] = role.DisplayName
	}

	return paramLabels
}

func (a *Templates) getRoleResources(role app.AppAWSIAMRoleConfig, t tagBuilder) map[string]cloudformation.Resource {
	rsrcs := make(map[string]cloudformation.Resource, 0)
	managedPolicyARNs := make([]string, 0)
	for _, policy := range role.Policies {
		if policy.ManagedPolicyName == "" {
			continue
		}

		managedPolicyARNs = append(managedPolicyARNs, fmt.Sprintf("arn:aws:iam::aws:policy/%s", policy.ManagedPolicyName))
	}

	trustPolicies := make([]map[string]any, 0)

	trustPolicies = append(trustPolicies, map[string]any{
		"Effect": "Allow",
		"Principal": map[string]any{
			"AWS": cloudformation.GetAttPtr("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRoleARN"),
		},
		"Action": "sts:AssumeRole",
	})

	if a.cfg.RunnerEnableSupport {
		trustPolicies = append(trustPolicies, map[string]any{
			"Effect": "Allow",
			"Principal": map[string]any{
				"AWS": []string{
					a.cfg.RunnerDefaultSupportIAMRole,
				},
			},
			"Action": "sts:AssumeRole",
		})
	}

	roleRsrc := &iam.Role{
		AWSCloudFormationCondition: a.roleConditionName(role),
		RoleName:                   generics.ToPtr(role.Name),
		ManagedPolicyArns:          managedPolicyARNs,
		AssumeRolePolicyDocument: map[string]any{
			"Statement": trustPolicies,
		},
		Tags: t.apply(nil, fmt.Sprintf("%s-role", role.Type)),
	}

	if len(role.PermissionsBoundaryJSON) < 1 || bytes.Equal(role.PermissionsBoundaryJSON, []byte("{}")) {
		// json.RawMessage requires the input to be valid json,
		// {} is a valid json but bytes slice with 0 size is not
		// role.PermissionsBoundaryJSON = []byte(`""`)
	} else {
		boundaryPolicyName := role.CloudFormationStackName + "Boundary"
		rsrcs[boundaryPolicyName] = a.getPermissionsBoundaryPolicy(role)

		roleRsrc.PermissionsBoundary = cloudformation.RefPtr(boundaryPolicyName)
	}

	// Add the role resource
	rsrcs[role.CloudFormationStackName] = roleRsrc

	// create each policy
	for _, policy := range role.Policies {
		if policy.ManagedPolicyName != "" {
			continue
		}

		rsrcs[policy.CloudFormationStackName] = a.getRolePolicy(role, policy)
	}

	return rsrcs
}

func (a *Templates) getRolesResources(inp *TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	rsrcs := make(map[string]cloudformation.Resource, 0)

	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		resources := a.getRoleResources(role, t)
		maps.Copy(rsrcs, resources)
	}

	for _, role := range inp.AppCfg.BreakGlassConfig.Roles {
		resources := a.getRoleResources(role, t)
		maps.Copy(rsrcs, resources)
	}

	return rsrcs
}

func (a *Templates) getPermissionsBoundaryPolicy(role app.AppAWSIAMRoleConfig) cloudformation.Resource {
	return &iam.ManagedPolicy{
		AWSCloudFormationCondition: a.roleConditionName(role),
		ManagedPolicyName:          generics.ToPtr(role.Name + "-boundary"),
		PolicyDocument:             json.RawMessage([]byte(role.PermissionsBoundaryJSON)),
	}
}

func (a *Templates) getRolePolicy(role app.AppAWSIAMRoleConfig, policy app.AppAWSIAMPolicyConfig) cloudformation.Resource {
	return &iam.Policy{
		AWSCloudFormationCondition: a.roleConditionName(role),
		PolicyName: cloudformation.SubVars(
			fmt.Sprintf("%s", policy.Name),
			map[string]any{"RoleName": cloudformation.Ref(role.CloudFormationStackName)}),
		PolicyDocument: json.RawMessage([]byte(policy.Contents)),
		Roles:          []string{cloudformation.Ref(role.CloudFormationStackName)},
	}
}

func (a *Templates) getRoleParameters(role app.AppAWSIAMRoleConfig) cloudformation.Parameter {
	return cloudformation.Parameter{
		Type:    "String",
		Default: true,
		AllowedValues: []any{
			"true",
			"false",
		},
		Description: generics.ToPtr(role.Description),
	}
}

func (a *Templates) getRolesParameters(inp *TemplateInput) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter, 0)

	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		params[role.CloudFormationStackParamName] = a.getRoleParameters(role)
	}
	for _, role := range inp.AppCfg.BreakGlassConfig.Roles {
		params[role.CloudFormationStackParamName] = a.getRoleParameters(role)
	}

	return params
}
