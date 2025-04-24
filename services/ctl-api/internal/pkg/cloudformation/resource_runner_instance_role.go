package cloudformation

import "github.com/awslabs/goformation/v7/cloudformation/iam"

func (a *Templates) getRunnerInstanceRole(inp *TemplateInput, t tagBuilder) *iam.Role {
	trustPolicy := []map[string]any{
		{
			"Effect": "Allow",
			"Principal": map[string]any{
				"Service": "ec2.amazonaws.com",
			},
			"Action": "sts:AssumeRole",
		},
	}

	return &iam.Role{
		Description: ptr("Instance role for the runner ec2 instance and ASG. Used to assume Provision, Deprovision, and Maintenance roles as needed by the app."),
		AssumeRolePolicyDocument: map[string]any{
			"Statement": trustPolicy,
		},
		Policies: []iam.Role_Policy{
			{
				PolicyName: "RunnerInstancePolicy",
				PolicyDocument: map[string]any{
					"Version": "2012-10-17",
					"Statement": []map[string]any{
						{
							"Effect": "Allow",
							"Action": []string{
								"sts:AssumeRole",
							},
							"Resource": "*",
						},
					},
				},
			},
		},
		Tags: t.apply(nil, "runner-instance"),
	}
}
