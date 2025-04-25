package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"

	"github.com/awslabs/goformation/v7/cloudformation/tags"
)

func (a *Templates) getRunnerSecurityGroup(inp *TemplateInput, t tagBuilder) *ec2.SecurityGroup {
	// NOTE: this tag is REQUIRED. the sandboxes use it to identify the runner group which
	// needs to be added to the eks node sg additional rules. this is a VIP tag. w/out it the
	// runner won't be allowed to apply heml/kubectl to the cluster.
	tags := []tags.Tag{
		{Key: "networking.nuon.co/domain", Value: "runner"},
	}
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
		Tags: t.apply(tags, "egress-sg"),
	}
}
