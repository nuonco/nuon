package cloudformation

import (
	"maps"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/iancoleman/strcase"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) getAWSTemplate(inp *stacks.TemplateInput) (*cloudformation.Template, error) {
	tmpl := cloudformation.NewTemplate()

	tb := tagBuilder{
		installID:  inp.Install.ID,
		additional: generics.ToStringMap(inp.Settings.AWSTags),
	}

	// build nested resources
	stack, vpcParams, err := t.getVPCNestedStack(inp, tb)
	if err != nil {
		return nil, err
	}
	tmpl.Resources["VPC"] = stack
	// vpcParams := t.getVPCNestedStackParams(inp)
	maps.Copy(tmpl.Parameters, vpcParams)

	// NOTE(fd): we branch here to choose the cf stack based on
	// NOTE(fd): this uses the configurable nested runner asg cf stack
	runnerASG, err := t.getRunnerASGNestedStack(inp, tb)
	if err != nil {
		return nil, err
	}
	tmpl.Resources["RunnerAutoScalingGroup"] = runnerASG

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

	// build roles (before custom nested stacks so they can depend on them)
	roles := t.getRolesResources(inp, tb)
	maps.Copy(tmpl.Resources, roles)
	roleParams := t.getRolesParameters(inp)
	maps.Copy(tmpl.Parameters, roleParams)
	roleConditions := t.getRoleConditions(inp)
	maps.Copy(tmpl.Conditions, roleConditions)
	roleParamLabels := t.getRolesParamLabels(inp)
	maps.Copy(paramlabels, roleParamLabels)

	// custom nested stacks
	existingResourceKeys := map[string]bool{}
	for k := range tmpl.Resources {
		existingResourceKeys[k] = true
	}
	customResult, err := t.getCustomNestedStacks(inp, tb, existingResourceKeys)
	if err != nil {
		return nil, err
	}
	for k, v := range customResult.resources {
		tmpl.Resources[k] = v
	}
	maps.Copy(tmpl.Parameters, customResult.params)

	// NOTE(fd): if there are no secrets in the config, the section is not rendered.
	// build secrets
	if len(inp.AppCfg.SecretsConfig.Secrets) > 0 {
		secrets := t.getSecretsResources(inp, tb)
		maps.Copy(tmpl.Resources, secrets)
		secretParams := t.getSecretsParameters(inp)
		maps.Copy(tmpl.Parameters, secretParams)
		secretConditions := t.getSecretsConditions(inp)
		maps.Copy(tmpl.Conditions, secretConditions)
		secretParamLabels := t.getSecretsParamLabels(inp)
		maps.Copy(paramlabels, secretParamLabels)
	}

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
	paramGroups := []map[string]any{
		{
			"Label": map[string]any{
				"default": "VPC Configuration",
			},
			"Parameters": pkggenerics.MapToKeys(vpcParams),
		},
	}
	if len(inp.AppCfg.SecretsConfig.Secrets) > 0 {
		paramGroups = append(paramGroups, map[string]any{
			"Label": map[string]any{
				"default": "Application Secrets",
			},
			"Parameters": pkggenerics.MapToKeys(t.getSecretsParameters(inp)),
		})
	}
	paramGroups = append(paramGroups, map[string]any{
		"Label": map[string]any{
			"default": "Access Permissions",
		},
		"Parameters": pkggenerics.MapToKeys(t.getRolesParameters(inp)),
	})
	pgs = append(pgs, paramGroups...)

	// add app input parameter group if there are any install_stack sourced inputs
	for groupName, installGroupParameters := range installGroupParameters {
		pgs = append(pgs, map[string]any{
			"Label": map[string]any{
				"default": "Install Inputs: " + strcase.ToCamel(groupName),
			},
			"Parameters": pkggenerics.MapToKeys(installGroupParameters),
		})
	}

	// add custom nested stack parameter groups
	pgs = append(pgs, customResult.paramGroups...)

	tmpl.Metadata["AWS::CloudFormation::Interface"] = map[string]any{
		"ParameterLabels": paramlabels,
		"ParameterGroups": pgs,
	}

	return tmpl, nil
}

func ptr[T any](v T) *T {
	return &v
}
