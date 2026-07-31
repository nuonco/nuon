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

func TestRunnerAuditExportParameterAndCondition(t *testing.T) {
	tpl := &Templates{}
	parameter := tpl.getRunnerAuditExportParameters()[runnerAuditExportConfigParameter]

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
	assert.Equal(t, "Runner Audit Export Configuration (Base64-encoded YAML)", tpl.getRunnerAuditExportParamLabels()[runnerAuditExportConfigParameter])

	assert.Equal(t,
		cloudformation.Not([]string{cloudformation.Equals(cloudformation.Ref(runnerAuditExportConfigParameter), "")}),
		tpl.getRunnerAuditExportConditions()[runnerAuditExportCondition],
	)
}

func TestRunnerAuditExportResources(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := &stacks.TemplateInput{
		Install:                    &app.Install{ID: "inst-test"},
		AppCfg:                     &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{},
		Settings:                   &app.RunnerGroupSettings{},
	}
	resources := tpl.getRunnerAuditExportResources(inp, tagBuilder{installID: inp.Install.ID})

	secret, ok := resources[runnerAuditExportSecret].(*secretsmanager.Secret)
	require.True(t, ok)
	assert.Equal(t, "nuon/inst-test/runner-audit-export", *secret.Name)
	assert.Equal(t, cloudformation.Ref(runnerAuditExportConfigParameter), *secret.SecretString)
	assert.Equal(t, runnerAuditExportCondition, secret.AWSCloudFormationCondition)

	policy, ok := resources["RunnerAuditExportSecretPolicy"].(*iam.Policy)
	require.True(t, ok)
	assert.Empty(t, policy.AWSCloudFormationCondition)
	assert.Equal(t, []string{cloudformation.GetAtt("RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole")}, policy.Roles)

	statement := policy.PolicyDocument.(map[string]interface{})["Statement"].([]interface{})[0].(map[string]interface{})
	assert.ElementsMatch(t, []string{"secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"}, statement["Action"])
	assert.Equal(t, []interface{}{cloudformation.Sub("arn:${AWS::Partition}:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:nuon/inst-test/runner-audit-export-*")}, statement["Resource"])
}

func TestRunnerAuditExportIsNotAddedToPhoneHome(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := &stacks.TemplateInput{
		Install:                    &app.Install{ID: "inst-test"},
		AppCfg:                     &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{},
		Settings:                   &app.RunnerGroupSettings{},
	}

	phoneHomeJSON, err := json.Marshal(tpl.getRunnerPhoneHomeProps(inp, nil))
	require.NoError(t, err)
	assert.NotContains(t, string(phoneHomeJSON), runnerAuditExportConfigParameter)
	assert.NotContains(t, string(phoneHomeJSON), runnerAuditExportSecret)
}
