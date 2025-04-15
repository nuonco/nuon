package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
)

func (a *Templates) getRunnerLaunchTemplatea(inp *TemplateInput, t tagBuilder) *ec2.LaunchTemplate {
	return &ec2.LaunchTemplate{
		LaunchTemplateName: ptr(cloudformation.Sub("${AWS::StackName}-runner")),
		LaunchTemplateData: a.getRunnerLaunchTemplateData(inp, t),
	}
}
