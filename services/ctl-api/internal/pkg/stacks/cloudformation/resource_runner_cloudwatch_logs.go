package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/logs"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (a *Templates) getRunnerCloudWatchLogGroup(inp *stacks.TemplateInput, t tagBuilder) *logs.LogGroup {
	return &logs.LogGroup{
		LogGroupName:    ptr(fmt.Sprintf("runner-%s", inp.Runner.ID)),
		RetentionInDays: ptr(7),
		Tags:            t.apply(nil, "runner-cw-lg"),
	}
}

func (a *Templates) getRunnerCloudWatchLogStream(inp *stacks.TemplateInput, t tagBuilder) *logs.LogStream {
	return &logs.LogStream{
		LogGroupName:  cloudformation.Ref("RunnerCloudWatchLogGroup"),
		LogStreamName: ptr(fmt.Sprintf("runner-%s", inp.Runner.ID)),
	}
}

func (a *Templates) getRunnerCloudWatchLogPolicy(inp *stacks.TemplateInput, t tagBuilder) *iam.Policy {
	return &iam.Policy{
		PolicyName: fmt.Sprintf("nuon-install-%s-cw-logs-access", inp.Install.ID),
		Roles: []string{
			cloudformation.GetAtt("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole"),
		},
		PolicyDocument: map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": []string{
						"logs:CreateLogStream",
						"logs:PutLogEvents",
					},
					"Effect": "Allow",
					"Resource": []interface{}{
						cloudformation.GetAtt("RunnerCloudWatchLogGroup", "Arn"),
					},
				},
			},
		},
	}
}
