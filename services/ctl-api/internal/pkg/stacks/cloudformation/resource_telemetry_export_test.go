package cloudformation

import (
	"encoding/json"
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func TestTelemetryExportParameterAndCondition(t *testing.T) {
	tpl := &Templates{}
	parameter := tpl.getTelemetryExportParameters()[telemetryExportConfigParameter]

	assert.Equal(t, "String", parameter.Type)
	require.NotNil(t, parameter.NoEcho)
	assert.True(t, *parameter.NoEcho)
	require.NotNil(t, parameter.MaxLength)
	assert.Equal(t, 4096, *parameter.MaxLength)
	defaultValue, ok := parameter.Default.(*string)
	require.True(t, ok)
	assert.Empty(t, *defaultValue)
	require.NotNil(t, parameter.Description)
	assert.Contains(t, *parameter.Description, "base64-encoded YAML")
	assert.Contains(t, *parameter.Description, "exporters.otlphttp.endpoint")
	assert.Contains(t, *parameter.Description, "optional headers")
	assert.Contains(t, *parameter.Description, "current release exports runner audit logs")
	assert.Equal(t, "Telemetry Export Configuration (Base64-encoded YAML)", tpl.getTelemetryExportParamLabels()[telemetryExportConfigParameter])

	assert.Equal(t,
		cloudformation.Not([]string{cloudformation.Equals(cloudformation.Ref(telemetryExportConfigParameter), "")}),
		tpl.getTelemetryExportConditions()[telemetryExportCondition],
	)
}

func TestTelemetryExportResources(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := &stacks.TemplateInput{
		Install:                    &app.Install{ID: "inst-test"},
		AppCfg:                     &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{},
		Settings:                   &app.RunnerGroupSettings{},
	}
	resources := tpl.getTelemetryExportResources(inp, tagBuilder{installID: inp.Install.ID})

	secret, ok := resources[telemetryExportSecret].(*secretsmanager.Secret)
	require.True(t, ok)
	assert.Equal(t, "nuon/inst-test/telemetry-export-config", *secret.Name)
	assert.Equal(t, cloudformation.Ref(telemetryExportConfigParameter), *secret.SecretString)
	assert.Equal(t, telemetryExportCondition, secret.AWSCloudFormationCondition)

	policy, ok := resources["TelemetryExportSecretPolicy"].(*iam.Policy)
	require.True(t, ok)
	assert.Empty(t, policy.AWSCloudFormationCondition)
	assert.Equal(t, []string{cloudformation.GetAtt("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole")}, policy.Roles)

	statement := policy.PolicyDocument.(map[string]interface{})["Statement"].([]interface{})[0].(map[string]interface{})
	assert.ElementsMatch(t, []string{"secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"}, statement["Action"])
	assert.Equal(t, []interface{}{cloudformation.Sub("arn:${AWS::Partition}:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:nuon/inst-test/telemetry-export-config-*")}, statement["Resource"])
}

func TestTelemetryExportIsNotAddedToPhoneHome(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := &stacks.TemplateInput{
		Install:                    &app.Install{ID: "inst-test"},
		AppCfg:                     &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{},
		Settings:                   &app.RunnerGroupSettings{},
	}

	phoneHomeJSON, err := json.Marshal(tpl.getRunnerPhoneHomeProps(inp, nil))
	require.NoError(t, err)
	assert.NotContains(t, string(phoneHomeJSON), telemetryExportConfigParameter)
	assert.NotContains(t, string(phoneHomeJSON), telemetryExportSecret)
}
