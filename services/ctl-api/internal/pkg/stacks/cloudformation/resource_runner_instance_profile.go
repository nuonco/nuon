package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (a *Templates) getRunnerInstanceProfile(inp *stacks.TemplateInput, t tagBuilder) *iam.InstanceProfile {
	return &iam.InstanceProfile{
		InstanceProfileName: ptr(cloudformation.Sub("${AWS::StackName}--runner-profile")),
		Roles: []string{
			cloudformation.Ref("RunnerInstanceRole"),
		},
	}
}
