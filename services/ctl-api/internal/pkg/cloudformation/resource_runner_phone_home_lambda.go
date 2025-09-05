package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/lambda"
)

func (a *Templates) getRunnerPhoneHomeProps(inp *TemplateInput) *cloudformation.CustomResource {
	lambdaprops := map[string]any{
		"ServiceToken": cloudformation.GetAttPtr("RunnerPhoneHome", "Arn"),
		"url":          inp.CloudFormationStackVersion.PhoneHomeURL,

		// fields for the phone-home endpoint
		"phone_home_type":          "aws",
		"maintenance_iam_role_arn": cloudformation.GetAttPtr("RunnerMaintenance", "Arn"),
		"provision_iam_role_arn":   cloudformation.GetAttPtr("RunnerProvision", "Arn"),
		"deprovision_iam_role_arn": cloudformation.GetAttPtr("RunnerDeprovision", "Arn"),
		"runner_iam_role_arn":      cloudformation.GetAttPtr("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole"),

		// from the nested VPC Cloudformation Template (we want its outputs)
		"vpc_id":          cloudformation.GetAtt("VPC", "Outputs.VPC"),
		"runner_subnet":   cloudformation.GetAtt("VPC", "Outputs.RunnerSubnet"),
		"public_subnets":  cloudformation.GetAtt("VPC", "Outputs.PublicSubnets"),
		"private_subnets": cloudformation.GetAtt("VPC", "Outputs.PrivateSubnets"),

		// account and region details
		"account_id": cloudformation.RefPtr("AWS::AccountId"),
		"region":     cloudformation.RefPtr("AWS::Region"),
	}

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		lambdaprops[secret.Name+"_arn"] = cloudformation.RefPtr(secret.CloudFormationStackName)
	}

	return &cloudformation.CustomResource{
		Type:       "AWS::CloudFormation::CustomResource",
		Properties: lambdaprops,
	}
}

func (a *Templates) getRunnerPhoneHomeLambda(inp *TemplateInput, t tagBuilder) *lambda.Function {
	// This is going to be moved into a cloudformation stack template and split out, with parameters for the body
	return &lambda.Function{
		Handler:     ptr("index.lambda_handler"),
		Runtime:     ptr("python3.12"),
		Tags:        t.apply(nil, "phone-home-lambda"),
		Description: ptr("Notify the Nuon API of the stack state."),
		Code: &lambda.Function_Code{
			ZipFile: ptr(inp.PhonehomeScript),
		},
		Role: cloudformation.GetAtt("RunnerPhoneHomeRole", "Arn"),
	}
}

func (a *Templates) getRunnerPhoneHomeLambdaRole(inp *TemplateInput, t tagBuilder) *iam.Role {
	return &iam.Role{
		Tags: t.apply(nil, "phone-home-lambda"),
		AssumeRolePolicyDocument: map[string]any{
			"Statement": []map[string]any{
				{
					"Effect": "Allow",
					"Principal": map[string]any{
						"Service": "lambda.amazonaws.com",
					},
					"Action": "sts:AssumeRole",
				},
			},
		},
		Policies: []iam.Role_Policy{
			{
				PolicyName: "CloudwatchPolicy",
				PolicyDocument: map[string]any{
					"Version": "2012-10-17",
					"Statement": []map[string]any{
						{
							"Effect": "Allow",
							"Action": []string{
								"logs:CreateLogGroup",
								"logs:CreateLogStream",
								"logs:PutLogEvents",
							},
							"Resource": "*",
						},
					},
				},
			},
		},
		ManagedPolicyArns: []string{
			"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
		},
	}
}
