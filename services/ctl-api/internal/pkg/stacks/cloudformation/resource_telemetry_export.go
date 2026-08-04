package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/secretsmanager"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	telemetryExportSecret = "TelemetryExportSecret"
)

func (a *Templates) getTelemetryExportResources(inp *stacks.TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	secret := &secretsmanager.Secret{
		Name: generics.ToPtr(fmt.Sprintf("nuon/%s/telemetry-export-config", inp.Install.ID)),
		Tags: t.apply(nil, "telemetry-export-config"),
	}

	policy := &iam.Policy{
		PolicyName: fmt.Sprintf("nuon-install-%s-telemetry-export-config-access", inp.Install.ID),
		Roles: []string{
			cloudformation.GetAtt("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole"),
		},
		PolicyDocument: map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": []string{
						"secretsmanager:GetSecretValue",
						"secretsmanager:DescribeSecret",
					},
					"Effect": "Allow",
					"Resource": []interface{}{
						cloudformation.Sub(fmt.Sprintf("arn:${AWS::Partition}:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:nuon/%s/telemetry-export-config-*", inp.Install.ID)),
					},
				},
			},
		},
	}

	return map[string]cloudformation.Resource{
		telemetryExportSecret:         secret,
		"TelemetryExportSecretPolicy": policy,
	}
}
