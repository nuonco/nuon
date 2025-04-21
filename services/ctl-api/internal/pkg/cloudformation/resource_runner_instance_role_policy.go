package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
)

func (a *Templates) getRunnerInstanceRoleCloudWatchLogPolicy(inp *TemplateInput, t tagBuilder) *iam.Policy {
	return &iam.Policy{
		PolicyName: fmt.Sprintf("nuon-install-%s-metadata", inp.Install.ID),
		Roles: []string{
			cloudformation.Ref("RunnerInstanceRole"),
		},
		PolicyDocument: map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": []string{
						"ec2:DescribeTags",
					},
					"Effect": "Allow",
					// selected resource only supports the wildcard
					// otherwise, we'd use limits like: `fmt.Sprintf("arn:aws:ec2:%s:%s:instance/*", inp.Install.AWSAccount.Region, inp.Install.AWSAccount.ID)`
					"Resource": "*",
					"Condition": map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"aws:Ec2InstanceSourceVpc": cloudformation.GetAtt("VPC", "Outputs.VPC"),
						},
					},
				},
			},
		},
	}
}
