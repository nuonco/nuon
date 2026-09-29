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
	if inp.CustomStacksOnly {
		return t.getAWSCustomStacksOnlyTemplate(inp)
	}

	tmpl := cloudformation.NewTemplate()

	tb := tagBuilder{
		installID:  inp.Install.ID,
		orgID:      inp.Install.OrgID,
		appID:      inp.Install.AppID,
		additional: generics.ToStringMap(inp.Settings.AWSTags),
	}

	stack, vpcParams, vpcOutputs, err := t.getVPCNestedStack(inp, tb)
	if err != nil {
		return nil, err
	}
	tmpl.Resources["VPC"] = stack
	maps.Copy(tmpl.Parameters, vpcParams)

	runnerParams := t.getRunnerParameters(inp)
	maps.Copy(tmpl.Parameters, runnerParams)
	paramlabels := map[string]any{}

	tmpl.Resources["RunnerSecurityGroup"] = t.getRunnerSecurityGroup(inp, tb)

	telemetryEndpoint := ""
	if !t.cfg.UseLocalRunners {
		runnerASG, supportsTelemetryIngress, err := t.getRunnerASGNestedStack(inp, tb)
		if err != nil {
			return nil, err
		}
		tmpl.Resources["RunnerAutoScalingGroup"] = runnerASG

		if _, hasPrefixList := vpcOutputs["VpcIpv4PrefixListId"]; hasPrefixList && supportsTelemetryIngress {
			parameter := cloudformation.Parameter{
				Type:          "String",
				Description:   ptr("Provision an internal OTLP HTTP endpoint for this install (additional AWS charges apply)"),
				Default:       "true",
				AllowedValues: []any{"true", "false"},
			}
			runnerParams["EnableTelemetryIngress"] = parameter
			tmpl.Parameters["EnableTelemetryIngress"] = parameter
			tmpl.Conditions["TelemetryIngressEnabled"] = cloudformation.Equals(cloudformation.Ref("EnableTelemetryIngress"), "true")
			runnerASG.Parameters["EnableTelemetryIngress"] = cloudformation.Ref("EnableTelemetryIngress")
			runnerASG.Parameters["VpcId"] = cloudformation.GetAtt("VPC", "Outputs.VPC")
			runnerASG.Parameters["TelemetrySourcePrefixListId"] = cloudformation.GetAtt("VPC", "Outputs.VpcIpv4PrefixListId")
			telemetryEndpoint = cloudformation.If("TelemetryIngressEnabled",
				cloudformation.GetAtt("RunnerAutoScalingGroup", "Outputs.TelemetryEndpoint"), "")
		}

		tmpl.Resources["RunnerCloudWatchLogGroup"] = t.getRunnerCloudWatchLogGroup(inp, tb)
		tmpl.Resources["RunnerCloudWatchLogStream"] = t.getRunnerCloudWatchLogStream(inp, tb)
		tmpl.Resources["RunnerCloudWatchLogPolicy"] = t.getRunnerCloudWatchLogPolicy(inp, tb)
		maps.Copy(tmpl.Resources, t.getTelemetryExportResources(inp, tb))
	}

	roles := t.getRolesResources(inp, tb)
	maps.Copy(tmpl.Resources, roles)
	roleParams := t.getRolesParameters(inp)
	maps.Copy(tmpl.Parameters, roleParams)
	roleConditions := t.getRoleConditions(inp)
	maps.Copy(tmpl.Conditions, roleConditions)
	roleParamLabels := t.getRolesParamLabels(inp)
	maps.Copy(paramlabels, roleParamLabels)

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

	if err := validatePhoneHomeScript(inp.PhonehomeScript); err != nil {
		return nil, err
	}
	phoneHomeProps := t.getRunnerPhoneHomeProps(inp, customResult)
	phoneHomeProps.Properties["telemetry_endpoint"] = telemetryEndpoint
	tmpl.Resources["PhoneHomeProps"] = phoneHomeProps
	tmpl.Outputs["TelemetryEndpoint"] = cloudformation.Output{
		Description: ptr("Private OTLP HTTP endpoint, empty when not provisioned"),
		Value:       telemetryEndpoint,
	}
	tmpl.Resources["RunnerPhoneHome"] = t.getRunnerPhoneHomeLambda(inp, tb)
	tmpl.Resources["RunnerPhoneHomeRole"] = t.getRunnerPhoneHomeLambdaRole(inp, tb)

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

	installGroupParameters := t.getInstallInputGroupParameters(inp)
	for _, installGroupParameter := range installGroupParameters {
		maps.Copy(tmpl.Parameters, installGroupParameter)
	}
	installGroupInputParamLables := t.getInstallInputGroupParamLable(inp)
	for _, installGroupParameLables := range installGroupInputParamLables {
		maps.Copy(paramlabels, installGroupParameLables)
	}

	var pgs []map[string]any
	paramGroups := []map[string]any{
		{
			"Label": map[string]any{
				"default": "Runner Configuration",
			},
			"Parameters": pkggenerics.MapToKeys(runnerParams),
		},
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

	for groupName, installGroupParameters := range installGroupParameters {
		pgs = append(pgs, map[string]any{
			"Label": map[string]any{
				"default": "Install Inputs: " + strcase.ToCamel(groupName),
			},
			"Parameters": pkggenerics.MapToKeys(installGroupParameters),
		})
	}

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
