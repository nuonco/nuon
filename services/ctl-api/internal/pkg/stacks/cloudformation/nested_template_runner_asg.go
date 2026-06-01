package cloudformation

import (
	"fmt"
	"slices"

	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"

	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// TODO: there are some runner env vars we want to provide defaults for IFF
// the user provides a value for them in `env_vars` block in the `runner.toml`.
func (a *Templates) getRunnerEnvVars(inp *stacks.TemplateInput) (string, error) {
	return inp.RunnerEnvVars, nil
}

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
			Value: a.cfg.RunnerAPIURL,
		},
	}

	params := map[string]string{
		"SubnetId":            cloudformation.GetAtt("VPC", "Outputs.RunnerSubnet"),
		"RunnerEgressGroupId": cloudformation.Ref("RunnerSecurityGroup"),
		"InstallId":           inp.Install.ID,
		"RunnerId":            inp.Runner.ID,
		"RunnerApiUrl":        a.cfg.RunnerAPIURL,
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

// getRunnerParameters returns the user-overridable top-level parameters for
// the runner. These are exposed at the parent stack so customers can override
// the defaults when creating/updating the stack; the nested runner ASG stack
// references them via Ref().
func (a *Templates) getRunnerParameters(inp *stacks.TemplateInput) map[string]cloudformation.Parameter {
	instanceType := inp.Settings.AWSInstanceType
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

	return map[string]cloudformation.Parameter{
		"RunnerInstanceType": {
			Type:          "String",
			Description:   generics.ToPtr("EC2 instance type for the runner"),
			Default:       instanceType,
			AllowedValues: allowedInstanceTypes,
		},
		"RunnerRootVolumeSize": {
			Type:        "Number",
			Description: generics.ToPtr("Root EBS volume size (GiB) for the runner"),
			Default:     "30",
			MinValue:    ptr(8.0),
			MaxValue:    ptr(100.0),
		},
	}
}
