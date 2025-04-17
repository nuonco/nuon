package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (t *Templates) getAWSEKSTemplate(inp *TemplateInput) (*cloudformation.Template, error) {
	tmpl := cloudformation.NewTemplate()

	tb := tagBuilder{
		installID:  inp.Install.ID,
		additional: generics.ToStringMap(inp.Settings.AWSTags),
	}

	// build nested resources
	tmpl.Resources["VPC"] = t.getVPCNestedStack(inp, tb)
	vpcParams := t.getVPCNestedStackParams()
	for name, param := range vpcParams {
		tmpl.Parameters[name] = param
	}

	// runner ASG and launch template
	tmpl.Resources["PhoneHomeProps"] = t.getRunnerPhoneHomeProps(inp)
	tmpl.Resources["RunnerPhoneHome"] = t.getRunnerPhoneHomeLambda(inp, tb)
	tmpl.Resources["RunnerPhoneHomeRole"] = t.getRunnerPhoneHomeLambdaRole(inp, tb)

	tmpl.Resources["RunnerLaunchTemplate"] = t.getRunnerLaunchTemplatea(inp, tb)
	tmpl.Resources["RunnerSecurityGroup"] = t.getRunnerSecurityGroup(inp, tb)
	tmpl.Resources["RunnerASG"] = t.getRunnerASG(inp, tb)
	tmpl.Resources["RunnerInstanceProfile"] = t.getRunnerInstanceProfile(inp, tb)

	tmpl.Resources["RunnerInstanceRole"] = t.getRunnerInstanceRole(inp, tb)

	// build roles
	paramlabels := map[string]any{}
	roles := t.getRolesResources(inp, tb)
	for rsrcName, rsrc := range roles {
		tmpl.Resources[rsrcName] = rsrc
	}
	roleParams := t.getRolesParameters(inp)
	for name, param := range roleParams {
		tmpl.Parameters[name] = param
	}
	roleConditions := t.getRoleConditions(inp)
	for name, condition := range roleConditions {
		tmpl.Conditions[name] = condition
	}
	roleParamLabels := t.getRolesParamLabels(inp)
	for name, paramLabel := range roleParamLabels {
		paramlabels[name] = paramLabel
	}

	// build secrets
	secrets := t.getSecretsResources(inp, tb)
	for rsrcName, rsrc := range secrets {
		tmpl.Resources[rsrcName] = rsrc
	}
	secretParams := t.getSecretsParameters(inp)
	for name, param := range secretParams {
		tmpl.Parameters[name] = param
	}
	secretParamLabels := t.getSecretsParamLabels(inp)
	for name, paramLabel := range secretParamLabels {
		paramlabels[name] = paramLabel
	}

	// parameter groups
	var pgs []map[string]any
	pgs = append(pgs, []map[string]any{
		{
			"Label": map[string]any{
				"default": "VPC Configuration",
			},
			"Parameters": pkggenerics.MapToKeys(t.getVPCNestedStackParams()),
		},
		{
			"Label": map[string]any{
				"default": "Application Secrets",
			},
			"Parameters": pkggenerics.MapToKeys(t.getSecretsParameters(inp)),
		},
		{
			"Label": map[string]any{
				"default": "Access Permissions",
			},
			"Parameters": pkggenerics.MapToKeys(t.getRolesParameters(inp)),
		},
	}...)
	tmpl.Metadata["AWS::CloudFormation::Interface"] = map[string]any{
		"ParameterLabels": paramlabels,
		"ParameterGroups": pgs,
	}

	return tmpl, nil
}

func ptr[T any](v T) *T {
	return &v
}
