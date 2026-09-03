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
	assert.Nil(t, secret.SecretString)
	assert.Empty(t, secret.AWSCloudFormationCondition)

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

	phoneHomeJSON, err := json.Marshal(tpl.getRunnerPhoneHomeProps(inp, nil, nil))
	require.NoError(t, err)
	assert.NotContains(t, string(phoneHomeJSON), telemetryExportSecret)
}
