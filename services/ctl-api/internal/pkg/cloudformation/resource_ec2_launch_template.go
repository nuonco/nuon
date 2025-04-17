package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
)

func (a *Templates) getRunnerLaunchTemplateData(inp *TemplateInput, t tagBuilder) *ec2.LaunchTemplate_LaunchTemplateData {
	return &ec2.LaunchTemplate_LaunchTemplateData{
		InstanceType: ptr(inp.Settings.AWSInstanceType),
		ImageId:      ptr(cloudformation.Sub("{{resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64}}")),
		IamInstanceProfile: &ec2.LaunchTemplate_IamInstanceProfile{
			Name: cloudformation.RefPtr("RunnerInstanceProfile"),
		},
		NetworkInterfaces: []ec2.LaunchTemplate_NetworkInterface{
			{
				AssociatePublicIpAddress: ptr(true),
				DeviceIndex:              ptr(0),
				SubnetId:                 cloudformation.GetAttPtr("VPC", "Outputs.RunnerSubnet"),
				Groups: []string{
					cloudformation.Ref("RunnerSecurityGroup"),
				},
			},
		},
		TagSpecifications: []ec2.LaunchTemplate_TagSpecification{
			{
				ResourceType: ptr("instance"),
				Tags: t.apply([]tags.Tag{
					{
						Key:   "nuon_runner_id",
						Value: inp.Runner.ID,
					},
					{
						Key:   "nuon_runner_api_url",
						Value: inp.Settings.RunnerAPIURL,
					},
					{
						Key:   "nuon_runner_api_token",
						Value: inp.APIToken,
					},
				}, "runner-instance"),
			},
			{
				ResourceType: ptr("network-interface"),
				Tags:         t.apply(nil, "runner-eni"),
			},
		},
		// in the beginning, there was a curlbash
		UserData: cloudformation.Base64Ptr(`#!/bin/bash
curl https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh | bash
`),
	}
}
