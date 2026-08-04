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
	telemetryExportConfigParameter = "TelemetryExportConfig"
	telemetryExportCondition       = "TelemetryExportConfigProvided"
	telemetryExportSecret          = "TelemetryExportSecret"
)

func (a *Templates) getTelemetryExportParameters() map[string]cloudformation.Parameter {
	return map[string]cloudformation.Parameter{
		telemetryExportConfigParameter: {
			Type:        "String",
			NoEcho:      generics.ToPtr(true),
			Default:     generics.ToPtr(""),
			MaxLength:   generics.ToPtr(4096),
			Description: generics.ToPtr("Optional base64-encoded YAML telemetry export configuration, shaped like an OTEL Collector exporter block: exporters.otlphttp.endpoint with optional headers. The current release exports runner audit logs. Encode with: base64 < config.yaml | tr -d '\\n'. Leave empty to disable."),
		},
	}
}

func (a *Templates) getTelemetryExportConditions() map[string]any {
	return map[string]any{
		telemetryExportCondition: cloudformation.Not([]string{
			cloudformation.Equals(cloudformation.Ref(telemetryExportConfigParameter), ""),
		}),
	}
}

func (a *Templates) getTelemetryExportParamLabels() map[string]any {
	return map[string]any{
		telemetryExportConfigParameter: "Telemetry Export Configuration (Base64-encoded YAML)",
	}
}

func (a *Templates) getTelemetryExportResources(inp *stacks.TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	secret := &secretsmanager.Secret{
		Name:                       generics.ToPtr(fmt.Sprintf("nuon/%s/telemetry-export-config", inp.Install.ID)),
		SecretString:               generics.ToPtr(cloudformation.Ref(telemetryExportConfigParameter)),
		Tags:                       t.apply(nil, "telemetry-export-config"),
		AWSCloudFormationCondition: telemetryExportCondition,
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
