package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
)

func (a *Templates) getRunnerSecurityGroup(inp *TemplateInput, t tagBuilder) *ec2.SecurityGroup {
	return &ec2.SecurityGroup{
		GroupDescription: "Egress security group for the runner - allow all outbound traffic",
		VpcId:            ptr(cloudformation.GetAtt("VPC", "Outputs.VPC")),
		SecurityGroupEgress: []ec2.SecurityGroup_Egress{
			{
				CidrIp:     ptr("0.0.0.0/0"),
				FromPort:   ptr(-1),
				ToPort:     ptr(-1),
				IpProtocol: "-1",
			},
		},
		Tags: t.apply(nil, "egress-sg"),
	}
}
