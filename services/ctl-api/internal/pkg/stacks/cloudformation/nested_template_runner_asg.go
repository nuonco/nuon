package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"

	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/nuonco/nuon/pkg/generics"
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
			Value: a.cfg.RunnerAPIURL,
		},
	}

	params := map[string]string{
		"SubnetId":            cloudformation.GetAtt("VPC", "Outputs.RunnerSubnet"),
		"RunnerEgressGroupId": cloudformation.Ref("RunnerSecurityGroup"),
		"InstallId":           inp.Install.ID,
		"RunnerId":            inp.Runner.ID,
		"RunnerApiUrl":        a.cfg.RunnerAPIURL,
		"InstanceType":        "t3a.medium",            // NOTE(fd): this may be a good thing to make configurable later
		"RootVolumeSize":      "30",                    // NOTE(fd): this may be a good thing to make configurable later
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

	return &nestedcloudformation.Stack{
		Parameters: params,
		TemplateURL: cloudformation.Join("", []interface{}{
			inp.AppCfg.StackConfig.RunnerNestedTemplateURL,
		}),
		Tags: t.apply(stackTags, "runner"),
	}, nil
}

// VPCNestedStackParameters returns the parameters for the VPC nested stack
func (a *Templates) getRunnerASGNestedStackParams() map[string]cloudformation.Parameter {
	return map[string]cloudformation.Parameter{
		"SubnetId": {
			Type:        "AWS::EC2::Subnet::Id",
			Description: generics.ToPtr("The subnet on which the app will run within the selected VPC."),
		},
		"RunnerEgressGroupId": {
			Type:        "AWS::EC2::SecurityGroup::Id",
			Description: generics.ToPtr("The security group for the runner instance that allows outbound traffic."),
		},
		"InstallId": {
			Type:        "String",
			Description: generics.ToPtr("The install ID"),
		},
		"RunnerId": {
			Type:        "String",
			Description: generics.ToPtr("The runner ID"),
		},
		"RunnerApiUrl": {
			Type:        "String",
			Description: generics.ToPtr("API URL for the runner"),
			Default:     "https://runner.nuon.co",
		},
		"InstanceType": {
			Type:        "String",
			Description: generics.ToPtr("EC2 instance type for the runner"),
			Default:     "t3a.medium",
			AllowedValues: []interface{}{
				"t3.small",
				"t3.medium",
				"t3a.small",
				"t3a.medium",
				"t3a.large",
				"m5.large",
				"m5a.large",
			},
		},
		"RootVolumeSize": {
			Type:        "Number",
			Description: generics.ToPtr("EC2 instance type for the runner"),
			Default:     "30",
			MinValue:    ptr(8.0),
			MaxValue:    ptr(100.0),
		},
	}
}
