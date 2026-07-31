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
	runnerAuditExportConfigParameter = "RunnerAuditExportConfig"
	runnerAuditExportCondition       = "RunnerAuditExportConfigProvided"
	runnerAuditExportSecret          = "RunnerAuditExportSecret"
)

func (a *Templates) getRunnerAuditExportParameters() map[string]cloudformation.Parameter {
	return map[string]cloudformation.Parameter{
		runnerAuditExportConfigParameter: {
			Type:        "String",
			NoEcho:      generics.ToPtr(true),
			Default:     generics.ToPtr(""),
			MaxLength:   generics.ToPtr(4096),
			Description: generics.ToPtr("Optional base64-encoded YAML configuration for a runner audit-log exporter, shaped like an OTEL Collector exporter block: exporters.otlphttp.endpoint with optional headers. Encode with: base64 < config.yaml | tr -d '\\n'. Leave empty to disable."),
		},
	}
}

func (a *Templates) getRunnerAuditExportConditions() map[string]any {
	return map[string]any{
		runnerAuditExportCondition: cloudformation.Not([]string{
			cloudformation.Equals(cloudformation.Ref(runnerAuditExportConfigParameter), ""),
		}),
	}
}

func (a *Templates) getRunnerAuditExportParamLabels() map[string]any {
	return map[string]any{
		runnerAuditExportConfigParameter: "Runner Audit Export Configuration (Base64-encoded YAML)",
	}
}

func (a *Templates) getRunnerAuditExportResources(inp *stacks.TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	secret := &secretsmanager.Secret{
		Name:                       generics.ToPtr(fmt.Sprintf("nuon/%s/runner-audit-export", inp.Install.ID)),
		SecretString:               generics.ToPtr(cloudformation.Ref(runnerAuditExportConfigParameter)),
		Tags:                       t.apply(nil, "runner-audit-export"),
		AWSCloudFormationCondition: runnerAuditExportCondition,
	}

	policy := &iam.Policy{
		PolicyName: fmt.Sprintf("nuon-install-%s-runner-audit-export-access", inp.Install.ID),
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
						cloudformation.Sub(fmt.Sprintf("arn:${AWS::Partition}:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:nuon/%s/runner-audit-export-*", inp.Install.ID)),
					},
				},
			},
		},
	}

	return map[string]cloudformation.Resource{
		runnerAuditExportSecret:         secret,
		"RunnerAuditExportSecretPolicy": policy,
	}
}
