package cloudformation

import (
	"maps"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/iancoleman/strcase"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (t *Templates) getAWSTemplate(inp *TemplateInput) (*cloudformation.Template, error) {
	tmpl := cloudformation.NewTemplate()

	tb := tagBuilder{
		installID:  inp.Install.ID,
		additional: generics.ToStringMap(inp.Settings.AWSTags),
	}

	// build nested resources
	tmpl.Resources["VPC"] = t.getVPCNestedStack(inp, tb)
	vpcParams := t.getVPCNestedStackParams()
	maps.Copy(tmpl.Parameters, vpcParams)

	// NOTE(fd): this uses the configurable neste runner asg cf stack
	tmpl.Resources["RunnerAutoScalingGroup"] = t.getRunnerASGNestedStack(inp, tb)

	// runner ASG and launch template
	tmpl.Resources["PhoneHomeProps"] = t.getRunnerPhoneHomeProps(inp)
	tmpl.Resources["RunnerPhoneHome"] = t.getRunnerPhoneHomeLambda(inp, tb)
	tmpl.Resources["RunnerPhoneHomeRole"] = t.getRunnerPhoneHomeLambdaRole(inp, tb)

	tmpl.Resources["RunnerSecurityGroup"] = t.getRunnerSecurityGroup(inp, tb)

	// CloudWatch: logs
	tmpl.Resources["RunnerCloudWatchLogGroup"] = t.getRunnerCloudWatchLogGroup(inp, tb)
	tmpl.Resources["RunnerCloudWatchLogStream"] = t.getRunnerCloudWatchLogStream(inp, tb)
	tmpl.Resources["RunnerCloudWatchLogPolicy"] = t.getRunnerCloudWatchLogPolicy(inp, tb)

	paramlabels := map[string]any{}

	// build roles
	roles := t.getRolesResources(inp, tb)
	maps.Copy(tmpl.Resources, roles)
	roleParams := t.getRolesParameters(inp)
	maps.Copy(tmpl.Parameters, roleParams)
	roleConditions := t.getRoleConditions(inp)
	maps.Copy(tmpl.Conditions, roleConditions)
	roleParamLabels := t.getRolesParamLabels(inp)
	maps.Copy(paramlabels, roleParamLabels)

	// build secrets
	secrets := t.getSecretsResources(inp, tb)
	maps.Copy(tmpl.Resources, secrets)
	secretParams := t.getSecretsParameters(inp)
	maps.Copy(tmpl.Parameters, secretParams)
	secretParamLabels := t.getSecretsParamLabels(inp)
	maps.Copy(paramlabels, secretParamLabels)

	// build app input parameters for install_stack sourced inputs
	installGroupParameters := t.getInstallInputGroupParameters(inp)
	for _, installGroupParameter := range installGroupParameters {
		maps.Copy(tmpl.Parameters, installGroupParameter)
	}
	installGroupInputParamLables := t.getInstallInputGroupParamLable(inp)
	for _, installGroupParameLables := range installGroupInputParamLables {
		maps.Copy(paramlabels, installGroupParameLables)
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

	// add app input parameter group if there are any install_stack sourced inputs
	for groupName, installGroupParameters := range installGroupParameters {
		pgs = append(pgs, map[string]any{
			"Label": map[string]any{
				"default": "Install Inputs: " + strcase.ToCamel(groupName),
			},
			"Parameters": pkggenerics.MapToKeys(installGroupParameters),
		})
	}

	tmpl.Metadata["AWS::CloudFormation::Interface"] = map[string]any{
		"ParameterLabels": paramlabels,
		"ParameterGroups": pgs,
	}

	return tmpl, nil
}

func ptr[T any](v T) *T {
	return &v
}
