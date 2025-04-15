package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/autoscaling"
)

func (a *Templates) getRunnerASG(inp *TemplateInput, t tagBuilder) *autoscaling.AutoScalingGroup {
	return &autoscaling.AutoScalingGroup{
		VPCZoneIdentifier: []string{},
		LaunchTemplate: &autoscaling.AutoScalingGroup_LaunchTemplateSpecification{
			LaunchTemplateId: cloudformation.RefPtr("RunnerLaunchTemplate"),
			Version:          cloudformation.GetAtt("RunnerLaunchTemplate", "LatestVersionNumber"),
		},
		MaxSize: "1",
		MinSize: "1",
		Tags: []autoscaling.AutoScalingGroup_TagProperty{
			{
				Key:               "nuon_install_id",
				Value:             inp.Install.ID,
				PropagateAtLaunch: false, // handled directly
			},
			{
				Key:               "Name",
				Value:             cloudformation.Sub("${AWS::StackName}-runner"),
				PropagateAtLaunch: false,
			},
		},
	}
}
