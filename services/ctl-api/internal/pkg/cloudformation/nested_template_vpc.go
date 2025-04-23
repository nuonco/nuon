package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"

	"github.com/powertoolsdev/mono/pkg/generics"
)

// VPCNestedStack returns a nested stack template for VPC resources
func (a *Templates) getVPCNestedStack(inp *TemplateInput, t tagBuilder) *nestedcloudformation.Stack {
	return &nestedcloudformation.Stack{
		Parameters: map[string]string{
			"VpcCIDR":           cloudformation.Ref("VpcCIDR"),
			"PublicSubnet1CIDR": cloudformation.Ref("PublicSubnet1CIDR"),
			"PublicSubnet2CIDR": cloudformation.Ref("PublicSubnet2CIDR"),
			"PublicSubnet3CIDR": cloudformation.Ref("PublicSubnet3CIDR"),

			"PrivateSubnet1CIDR": cloudformation.Ref("PrivateSubnet1CIDR"),
			"PrivateSubnet2CIDR": cloudformation.Ref("PrivateSubnet2CIDR"),
			"PrivateSubnet3CIDR": cloudformation.Ref("PrivateSubnet3CIDR"),

			"RunnerSubnetCIDR": cloudformation.Ref("RunnerSubnetCIDR"),

			"ClusterName":   inp.Install.ID,
			"NuonInstallID": inp.Install.ID,
			"NuonAppID":     inp.Install.AppID,
			"NuonOrgID":     inp.Install.OrgID,
		},
		TemplateURL: cloudformation.Join("", []interface{}{
			inp.AppCfg.StackConfig.VPCNestedTemplateURL,
		}),
		Tags: t.apply(nil, "vpc"),
	}
}

// VPCNestedStackParameters returns the parameters for the VPC nested stack
func (a *Templates) getVPCNestedStackParams() map[string]cloudformation.Parameter {
	return map[string]cloudformation.Parameter{
		"VpcCIDR": {
			Type:        "String",
			Default:     "10.128.0.0/16",
			Description: generics.ToPtr("CIDR block for the VPC"),
		},

		// private subnet cidrs
		"PrivateSubnet1CIDR": {
			Type:        "String",
			Default:     "10.128.130.0/24",
			Description: generics.ToPtr("CIDR block for private subnet 1"),
		},
		"PrivateSubnet2CIDR": {
			Type:        "String",
			Default:     "10.128.132.0/24",
			Description: generics.ToPtr("CIDR block for private subnet 2"),
		},
		"PrivateSubnet3CIDR": {
			Type:        "String",
			Default:     "10.128.134.0/24",
			Description: generics.ToPtr("CIDR block for private subnet 3"),
		},

		// runner subnet cidr
		"RunnerSubnetCIDR": {
			Type:        "String",
			Default:     "10.128.128.0/24",
			Description: generics.ToPtr("CIDR block for the Runner Subnet"),
		},

		// public subnet cidr
		"PublicSubnet1CIDR": {
			Type:        "String",
			Default:     "10.128.0.0/26",
			Description: generics.ToPtr("CIDR block for public subnet 1"),
		},
		"PublicSubnet2CIDR": {
			Type:        "String",
			Default:     "10.128.0.64/26",
			Description: generics.ToPtr("CIDR block for public subnet 2"),
		},
		"PublicSubnet3CIDR": {
			Type:        "String",
			Default:     "10.128.0.128/26",
			Description: generics.ToPtr("CIDR block for public subnet 3"),
		},
	}
}
