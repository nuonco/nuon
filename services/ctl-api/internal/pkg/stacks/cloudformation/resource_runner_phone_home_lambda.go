package cloudformation

import (
	"fmt"

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

// lambdaInlineCodeLimit is AWS's hard cap on Code.ZipFile, the inline source the
// phone-home Lambda uses.
const lambdaInlineCodeLimit = 4096

// validatePhoneHomeScript fails the render rather than letting an empty or oversized
// script fail at CreateStack inside the customer's account, where the error is far
// harder to attribute. The escape hatch when the script legitimately outgrows this
// is an S3-hosted zip (Code.S3Bucket/S3Key) in the existing template bucket.
//
// Empty is checked here rather than as a `validate:"required"` tag on TemplateInput
// because the Azure and GCP renderers share that struct and have no phone-home Lambda.
func validatePhoneHomeScript(script string) error {
	if script == "" {
		return fmt.Errorf("phone home script is empty")
	}

	if len(script) > lambdaInlineCodeLimit {
		return fmt.Errorf(
			"phone home script is %d bytes, over the %d byte limit for inline lambda source",
			len(script), lambdaInlineCodeLimit,
		)
	}

	return nil
}

// Both of these were CloudFormation defaults (3s, 128MB) until the phone-home script
// started fetching a token, and both defaults were wrong for it.
//
// The timeout has to clear the script's own retry ladder — MAX_RETRIES=5 with
// BASE_DELAY=1.75 and exponential backoff is 26.25s of sleeps before it gives up — plus
// the requests either side of it. Under the 3s default the ladder was unreachable code
// that had never completed a single retry.
//
// Memory is a CPU setting here, not a memory one: Lambda scales vCPU with it, and 128MB
// is roughly a twelfth of one. The token fetch imports boto3 and builds a client, which
// loads botocore's service models, and at that CPU share it does not finish inside 3
// seconds — the function was killed before it ever called Secrets Manager, so the phone
// home never went out and the stack hung on the custom resource until it rolled back.
// 512MB is about 4x the CPU for roughly a quarter of the duration, so the cost is close
// to a wash.
const (
	phoneHomeLambdaTimeoutSeconds = 60
	phoneHomeLambdaMemoryMB       = 512
)

func (a *Templates) getRunnerPhoneHomeLambda(inp *stacks.TemplateInput, t tagBuilder) *lambda.Function {
	// This is going to be moved into a cloudformation stack template and split out, with parameters for the body
	fn := &lambda.Function{
		Handler:     ptr("index.lambda_handler"),
		Runtime:     ptr("python3.12"),
		Tags:        t.apply(nil, "phone-home-lambda"),
		Description: ptr("Notify the Nuon API of the stack state."),
		Timeout:     ptr(phoneHomeLambdaTimeoutSeconds),
		MemorySize:  ptr(phoneHomeLambdaMemoryMB),
		Code: &lambda.Function_Code{
			ZipFile: ptr(inp.PhonehomeScript),
		},
		Role: cloudformation.GetAtt("RunnerPhoneHomeRole", "Arn"),
	}

	// Environment variables rather than resource properties, deliberately.
	// phonehome.py does `props = data.pop("ResourceProperties")` and POSTs every
	// prop, so anything added to getRunnerPhoneHomeProps is echoed back into the
	// phone-home body and lands in InstallStackVersionRun.Data. Env vars are read
	// via os.environ and are never echoed.
	if inp.PhoneHomeSecretARN != "" {
		fn.Environment = &lambda.Function_Environment{
			Variables: map[string]string{
				"NUON_PHONE_HOME_SECRET_ARN": inp.PhoneHomeSecretARN,
				// The region of the *secret* — Nuon's management region, not the
				// install's. The Lambda has no VPC config, so a cross-region
				// Secrets Manager call is fine.
				"NUON_PHONE_HOME_SECRET_REGION": inp.PhoneHomeSecretRegion,
				// Which entry to read out of the token map. Scoped to the stack
				// version this template belongs to, so it can never go stale: a
				// template never needs a phone_home_id other than its own.
				"NUON_PHONE_HOME_ID": inp.CloudFormationStackVersion.PhoneHomeID,
			},
		}
	}

	return fn
}

func (a *Templates) getRunnerPhoneHomeLambdaRole(inp *stacks.TemplateInput, t tagBuilder) *iam.Role {
	role := a.basePhoneHomeLambdaRole(inp, t)

	// The identity half of the cross-account read. The secret's resource policy in
	// Nuon's management account names this role as a principal; this narrows what the
	// role itself may do. Both sides are required — either alone denies.
	if inp.PhoneHomeSecretARN != "" {
		statements := []map[string]any{
			{
				"Effect":   "Allow",
				"Action":   []string{"secretsmanager:GetSecretValue"},
				"Resource": inp.PhoneHomeSecretARN,
			},
		}

		// Scoped to the CMK that encrypted the secret. Without kms:Decrypt the read
		// fails no matter how permissive the resource policy is, because the
		// AWS-managed secretsmanager key cannot be shared cross-account.
		if arn := a.cfg.AWSPhoneHomeCMKARN; arn != "" {
			statements = append(statements, map[string]any{
				"Effect":   "Allow",
				"Action":   []string{"kms:Decrypt", "kms:DescribeKey"},
				"Resource": arn,
			})
		}

		role.Policies = append(role.Policies, iam.Role_Policy{
			PolicyName: "PhoneHomeSecretPolicy",
			PolicyDocument: map[string]any{
				"Version":   "2012-10-17",
				"Statement": statements,
			},
		})
	}

	return role
}

func (a *Templates) basePhoneHomeLambdaRole(inp *stacks.TemplateInput, t tagBuilder) *iam.Role {
	return &iam.Role{
		// Named so a cross-account grant can reference this principal before the
		// customer's stack has created it. Setting RoleName makes the role a
		// replacement on the next stack update for installs whose role currently has
		// a CloudFormation-generated name.
		RoleName: ptr(stacks.PhoneHomeRoleName(inp.Install.ID)),
		Tags:     t.apply(nil, "phone-home-lambda"),
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
