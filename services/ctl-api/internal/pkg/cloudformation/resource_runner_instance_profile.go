package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
)

func (a *Templates) getRunnerInstanceProfile(inp *TemplateInput, t tagBuilder) *iam.InstanceProfile {
	return &iam.InstanceProfile{
		InstanceProfileName: ptr(cloudformation.Sub("${AWS::StackName}--runner-profile")),
		Roles: []string{
			cloudformation.Ref("RunnerInstanceRole"),
		},
	}
}
