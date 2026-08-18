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

// getRunnerASGNestedStack returns a nested stack template for runner ASG resources.
// It fetches the runner template to discover its parameters, conditionally including
// RunnerApiToken only if the template defines it.
func (a *Templates) getRunnerASGNestedStack(inp *stacks.TemplateInput, t tagBuilder) (*nestedcloudformation.Stack, error) {
	// fetch the runner template to inspect its declared parameters
	tmpl, err := a.fetchTemplate(inp.AppCfg.StackConfig.RunnerNestedTemplateURL)
	if err != nil {
		return nil, fmt.Errorf("runner ASG nested stack: %w", err)
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
		"RunnerInitScriptUrl": inp.RunnerInitScriptURL, // NOTE(fd): this is user- (provided/configurable)
	}

	// conditionally include RunnerApiToken if the nested template defines it as a parameter
	if _, ok := tmpl.Parameters["RunnerApiToken"]; ok {
		params["RunnerApiToken"] = inp.APIToken
		stackTags = append(stackTags, tags.Tag{
			Key:   "nuon_runner_api_token",
			Value: inp.APIToken,
		})
	}

	// conditionally include RunnerEnvVars if the nested template defines it as a parameter
	if _, ok := tmpl.Parameters["RunnerEnvVars"]; ok {
		params["RunnerEnvVars"] = inp.RunnerEnvVars
	}

	return &nestedcloudformation.Stack{
		Parameters: params,
		TemplateURL: cloudformation.Join("", []interface{}{
			inp.AppCfg.StackConfig.RunnerNestedTemplateURL,
		}),
		Tags: t.apply(stackTags, "runner"),
	}, nil
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

// getRunnerParameters returns the user-overridable top-level parameters for
// the runner. These are exposed at the parent stack so customers can override
// the defaults when creating/updating the stack; the nested runner ASG stack
// references them via Ref().
//
// Defaults come from the nested runner template when it declares the matching
// parameter, so a customer template that ships its own sizing is not silently
// overridden by the platform defaults.
func (a *Templates) getRunnerParameters(inp *stacks.TemplateInput) map[string]cloudformation.Parameter {
	tmplParams := a.runnerTemplateParameters(inp)

	return map[string]cloudformation.Parameter{
		"RunnerInstanceType":   a.runnerInstanceTypeParameter(inp, tmplParams["InstanceType"]),
		"RunnerRootVolumeSize": a.runnerRootVolumeSizeParameter(tmplParams["RootVolumeSize"]),
	}
}

// runnerTemplateParameters returns the parameters declared by the runner nested
// template. It yields nil when the template cannot be fetched or parsed so the
// platform defaults apply; getRunnerASGNestedStack surfaces the fetch error for
// every stack that actually deploys the runner ASG.
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

// The app's own runner config wins, then the nested template's declared default, then the
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

	// Always allow the configured instance type so a custom value from
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
	// the default has to satisfy the bounds, otherwise CloudFormation rejects the
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
