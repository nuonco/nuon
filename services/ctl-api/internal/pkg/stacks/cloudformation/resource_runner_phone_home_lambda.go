package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/lambda"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (a *Templates) getRunnerPhoneHomeProps(inp *stacks.TemplateInput, customStacks *customNestedStackResult) *cloudformation.CustomResource {
	breakGlassRoleArns := make(map[string]interface{})
	customRoleArns := make(map[string]interface{})

	for _, role := range inp.AppCfg.BreakGlassConfig.Roles {
		// cloudformation has parameter called role.CloudFormationStackParamName
		breakGlassRoleArns[role.Name] = cloudformation.If(
			role.CloudFormationStackParamName,
			generics.FromPtrStr(cloudformation.GetAttPtr(role.CloudFormationStackName, "Arn")),
			cloudformation.Ref("AWS::NoValue"),
		)
	}

	for _, role := range inp.AppCfg.PermissionsConfig.CustomRoles {
		customRoleArns[role.Name] = cloudformation.If(
			role.CloudFormationStackParamName,
			generics.FromPtrStr(cloudformation.GetAttPtr(role.CloudFormationStackName, "Arn")),
			cloudformation.Ref("AWS::NoValue"),
		)
	}

	// add app input parameters from install_stack sourced inputs
	installInputValues := make(map[string]interface{})
	for _, input := range inp.AppCfg.InputConfig.AppInputs {
		if input.Source == "customer" {
			// Use the original input name as the key in nested object
			installInputValues[input.Name] = cloudformation.RefPtr(input.CloudFormationStackParamName)
		}
	}

	// Build conditional role ARN references so that disabling a role
	// (e.g. EnableRunnerProvision=false) doesn't create an unresolvable
	// Fn::GetAtt on a resource that doesn't exist.
	roleArnByType := make(map[string]any)
	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		roleArnByType[string(role.Type)] = cloudformation.If(
			role.CloudFormationStackParamName,
			generics.FromPtrStr(cloudformation.GetAttPtr(role.CloudFormationStackName, "Arn")),
			"",
		)
	}

	lambdaprops := map[string]any{
		"ServiceToken": cloudformation.GetAttPtr("RunnerPhoneHome", "Arn"),
		"url":          inp.CloudFormationStackVersion.PhoneHomeURL,

		// fields for the phone-home endpoint
		"phone_home_type":          "aws",
		"maintenance_iam_role_arn": roleArnByType[string(app.AWSIAMRoleTypeRunnerMaintenance)],
		"provision_iam_role_arn":   roleArnByType[string(app.AWSIAMRoleTypeRunnerProvision)],
		"deprovision_iam_role_arn": roleArnByType[string(app.AWSIAMRoleTypeRunnerDeprovision)],

		"install_inputs":        installInputValues,
		"break_glass_role_arns": breakGlassRoleArns,
		"custom_role_arns":      customRoleArns,

		// from the nested VPC Cloudformation Template (we want its outputs)
		"vpc_id":          cloudformation.GetAtt("VPC", "Outputs.VPC"),
		"runner_subnet":   cloudformation.GetAtt("VPC", "Outputs.RunnerSubnet"),
		"public_subnets":  cloudformation.GetAtt("VPC", "Outputs.PublicSubnets"),
		"private_subnets": cloudformation.GetAtt("VPC", "Outputs.PrivateSubnets"),

		// account and region details
		"account_id": cloudformation.RefPtr("AWS::AccountId"),
		"region":     cloudformation.RefPtr("AWS::Region"),
	}

	// runner_iam_role_arn references the ASG nested stack output, only available
	// when not using local runners
	if !a.cfg.UseLocalRunners {
		lambdaprops["runner_iam_role_arn"] = cloudformation.GetAttPtr("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole")
	}

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		if secret.AutoGenerate || secret.Required {
			lambdaprops[secret.Name+"_arn"] = cloudformation.RefPtr(secret.CloudFormationStackName)
			continue
		}

		lambdaprops[secret.Name+"_arn"] = cloudformation.If(
			a.secretConditionName(secret.CloudFormationParamName),
			cloudformation.Ref(secret.CloudFormationStackName),
			"",
		)
	}

	// Always include custom_nested_stacks in the payload (empty map when none configured)
	customNestedStacksPayload := map[string]any{}
	if customStacks != nil {
		for logicalID, info := range customStacks.stackOutputs {
			outputsMap := map[string]any{}
			for _, outputKey := range info.OutputKeys {
				outputsMap[outputKey] = cloudformation.GetAtt(logicalID, "Outputs."+outputKey)
			}
			customNestedStacksPayload[info.Name] = map[string]any{
				"outputs": outputsMap,
			}
		}
	}
	lambdaprops["custom_nested_stacks"] = customNestedStacksPayload

	resource := &cloudformation.CustomResource{
		Type:       "AWS::CloudFormation::CustomResource",
		Properties: lambdaprops,
	}
	if customStacks != nil && customStacks.lastLogicalID != "" {
		resource.AWSCloudFormationDependsOn = []string{customStacks.lastLogicalID}
	}
	return resource
}

func (a *Templates) getRunnerPhoneHomeLambda(inp *stacks.TemplateInput, t tagBuilder) *lambda.Function {
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

func (a *Templates) getRunnerPhoneHomeLambdaRole(inp *stacks.TemplateInput, t tagBuilder) *iam.Role {
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
