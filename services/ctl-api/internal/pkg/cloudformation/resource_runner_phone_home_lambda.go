package cloudformation

import (
	"io"
	"net/http"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/lambda"
	"github.com/pkg/errors"
)

func (a *Templates) getRunnerPhoneHomeProps(inp *TemplateInput) *cloudformation.CustomResource {
	lambdaprops := map[string]any{
		"ServiceToken": cloudformation.GetAttPtr("RunnerPhoneHome", "Arn"),
		"url":          inp.CloudFormationStackVersion.PhoneHomeURL,

		// fields for the phone-home endpoint
		"phone_home_type":          "aws",
		"maintenance_iam_role_arn": "maintenance-role",
		"provision_iam_role_arn":   "provision-role",
		"deprovision_iam_role_arn": "deprovision-role",
		"account_id":               "account-id",
		"vpc_id":                   "vpc-id",
	}

	return &cloudformation.CustomResource{
		Type:       "AWS::CloudFormation::CustomResource",
		Properties: lambdaprops,
	}
}

func (a *Templates) getRunnerPhoneHomeLambda(inp *TemplateInput, t tagBuilder) *lambda.Function {
	// This is going to be moved into a cloudformation stack template and split out, with parameters for the body
	// Grab the latest version of the phone-home script
	resp, err := http.Get("https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/phonehome.py")
	if err != nil {
		panic(errors.Wrap(err, "failed to fetch phone-home script"))
	}
	byts, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(errors.Wrap(err, "failed to read body of phone-home script"))
	}

	return &lambda.Function{
		Handler:     ptr("index.lambda_handler"),
		Runtime:     ptr("python3.9"),
		Tags:        t.apply(nil, "phone-home-lambda"),
		Description: ptr("Notify the Nuon API of the stack state."),
		Code: &lambda.Function_Code{
			ZipFile: ptr(string(byts)),
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
