package cloudformation

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"

	"github.com/awslabs/goformation/v7/cloudformation/tags"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (a *Templates) getRunnerASGNestedStack(inp *stacks.TemplateInput, t tagBuilder) (*nestedcloudformation.Stack, bool, error) {
	tmpl, err := a.fetchTemplate(inp.AppCfg.StackConfig.RunnerNestedTemplateURL)
	if err != nil {
		return nil, false, fmt.Errorf("runner ASG nested stack: %w", err)
	}

	stackTags := []tags.Tag{
		{
			Key:   "Name",
			Value: fmt.Sprintf("%s-runner-instance", inp.Install.ID),
		},
		{
			Key:   "nuon_runner_id",
			Value: inp.Runner.ID,
		},
		{
			Key:   "runner.nuon.co/id",
			Value: inp.Runner.ID,
		},
		{
			Key:   "nuon_runner_api_url",
			Value: a.runnerAPIURL(inp),
		},
	}

	params := map[string]string{
		"SubnetId":            cloudformation.GetAtt("VPC", "Outputs.RunnerSubnet"),
		"RunnerEgressGroupId": cloudformation.Ref("RunnerSecurityGroup"),
		"InstallId":           inp.Install.ID,
		"RunnerId":            inp.Runner.ID,
		"RunnerApiUrl":        a.runnerAPIURL(inp),
		"InstanceType":        cloudformation.Ref("RunnerInstanceType"),
		"RootVolumeSize":      cloudformation.Ref("RunnerRootVolumeSize"),
		"RunnerInitScriptUrl": inp.RunnerInitScriptURL,
	}

	if _, ok := tmpl.Parameters["RunnerApiToken"]; ok {
		params["RunnerApiToken"] = inp.APIToken
		stackTags = append(stackTags, tags.Tag{
			Key:   "nuon_runner_api_token",
			Value: inp.APIToken,
		})
	}

	if _, ok := tmpl.Parameters["RunnerEnvVars"]; ok {
		params["RunnerEnvVars"] = inp.RunnerEnvVars
	}

	_, supportsTelemetryIngress := tmpl.Outputs["TelemetryEndpoint"]
	for _, name := range []string{"EnableTelemetryIngress", "VpcId", "TelemetrySourcePrefixListId"} {
		if _, ok := tmpl.Parameters[name]; !ok {
			supportsTelemetryIngress = false
		}
	}

	return &nestedcloudformation.Stack{
		Parameters: params,
		TemplateURL: cloudformation.Join("", []interface{}{
			inp.AppCfg.StackConfig.RunnerNestedTemplateURL,
		}),
		Tags: t.apply(stackTags, "runner"),
	}, supportsTelemetryIngress, nil
}

func (a *Templates) runnerAPIURL(inp *stacks.TemplateInput) string {
	if inp.Settings != nil && inp.Settings.RunnerAPIURL != "" {
		return inp.Settings.RunnerAPIURL
	}
	return a.cfg.RunnerAPIURL
}

const (
	defaultRunnerRootVolumeSize = 30.0
	minRunnerRootVolumeSize     = 8.0
	maxRunnerRootVolumeSize     = 100.0
)

func (a *Templates) getRunnerParameters(inp *stacks.TemplateInput) map[string]cloudformation.Parameter {
	tmplParams := a.runnerTemplateParameters(inp)

	return map[string]cloudformation.Parameter{
		"RunnerInstanceType":   a.runnerInstanceTypeParameter(inp, tmplParams["InstanceType"]),
		"RunnerRootVolumeSize": a.runnerRootVolumeSizeParameter(tmplParams["RootVolumeSize"]),
	}
}

func (a *Templates) runnerTemplateParameters(inp *stacks.TemplateInput) map[string]cfnParameterShape {
	if inp.AppCfg == nil || inp.AppCfg.StackConfig.RunnerNestedTemplateURL == "" {
		return nil
	}

	tmpl, err := a.fetchTemplate(inp.AppCfg.StackConfig.RunnerNestedTemplateURL)
	if err != nil {
		return nil
	}

	return tmpl.Parameters
}

// why: The app's own runner config wins, then the nested template's declared default, then the
// platform default. Settings.AWSInstanceType is deliberately not consulted: the stack
// generators resolve the platform default into it, so it is never empty and would mask the
// template's default entirely.
func (a *Templates) runnerInstanceTypeParameter(inp *stacks.TemplateInput, tmplParam cfnParameterShape) cloudformation.Parameter {
	instanceType := inp.ConfiguredRunnerInstanceType
	if instanceType == "" {
		instanceType, _ = tmplParam.Default.(string)
	}
	if instanceType == "" {
		instanceType = app.DefaultAWSInstanceType
	}

	// why: Always allow the configured instance type so a custom value from
	// runner.toml is never rejected by the AllowedValues constraint.
	allowedInstanceTypes := []interface{}{
		"t3.medium",
		"t3.large",
		"t3a.medium",
		"t3a.large",
		"c4.large",
		"c5.large",
	}
	if !slices.Contains(allowedInstanceTypes, interface{}(instanceType)) {
		allowedInstanceTypes = append(allowedInstanceTypes, instanceType)
	}

	return cloudformation.Parameter{
		Type:          "String",
		Description:   generics.ToPtr("EC2 instance type for the runner"),
		Default:       instanceType,
		AllowedValues: allowedInstanceTypes,
	}
}

func (a *Templates) runnerRootVolumeSizeParameter(tmplParam cfnParameterShape) cloudformation.Parameter {
	size := defaultRunnerRootVolumeSize
	if tmplDefault, ok := numericParamValue(tmplParam.Default); ok {
		size = tmplDefault
	}

	minSize, maxSize := minRunnerRootVolumeSize, maxRunnerRootVolumeSize
	if tmplParam.MinValue != nil {
		minSize = *tmplParam.MinValue
	}
	if tmplParam.MaxValue != nil {
		maxSize = *tmplParam.MaxValue
	}
	// why: the default has to satisfy the bounds, otherwise CloudFormation rejects the
	// parent template outright.
	minSize = math.Min(minSize, size)
	maxSize = math.Max(maxSize, size)

	return cloudformation.Parameter{
		Type:        "Number",
		Description: generics.ToPtr("Root EBS volume size (GiB) for the runner"),
		Default:     strconv.FormatFloat(size, 'f', -1, 64),
		MinValue:    ptr(minSize),
		MaxValue:    ptr(maxSize),
	}
}

func numericParamValue(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
