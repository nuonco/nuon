package cloudformation

import (
	"strings"
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func phoneHomeTestInput(installID string) *stacks.TemplateInput {
	return &stacks.TemplateInput{
		Install: &app.Install{ID: installID},
		AppCfg:  &app.AppConfig{},
		CloudFormationStackVersion: &app.InstallStackVersion{
			PhoneHomeURL: "https://example.com/phone-home",
		},
	}
}

const (
	testPhoneHomeSecretARN = "arn:aws:secretsmanager:us-west-2:123456789012:secret:nuon/phone-home/inst1-aB3xYz"
	testPhoneHomeCMKARN    = "arn:aws:kms:us-west-2:123456789012:key/abcd-1234"
)

func phoneHomeAuthTestInput(installID string) *stacks.TemplateInput {
	inp := phoneHomeTestInput(installID)
	inp.CloudFormationStackVersion.PhoneHomeID = "phv7g2k9x4m1qz8w3n6b5t0jrc"
	inp.PhoneHomeSecretARN = testPhoneHomeSecretARN
	inp.PhoneHomeSecretRegion = "us-west-2"

	return inp
}

func TestGetRunnerPhoneHomeLambda_SecretEnvVars(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	fn := tpl.getRunnerPhoneHomeLambda(inp, tagBuilder{installID: inp.Install.ID})

	require.NotNil(t, fn.Environment)
	assert.Equal(t, testPhoneHomeSecretARN, fn.Environment.Variables["NUON_PHONE_HOME_SECRET_ARN"])
	assert.Equal(t, "us-west-2", fn.Environment.Variables["NUON_PHONE_HOME_SECRET_REGION"])
	assert.Equal(t, "phv7g2k9x4m1qz8w3n6b5t0jrc", fn.Environment.Variables["NUON_PHONE_HOME_ID"])
}

func TestGetRunnerPhoneHomeLambda_NoEnvWithoutSecret(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	fn := tpl.getRunnerPhoneHomeLambda(inp, tagBuilder{installID: inp.Install.ID})

	assert.Nil(t, fn.Environment, "no phone-home secret means no environment block")
}

func TestGetRunnerPhoneHomeLambdaRole_SecretPolicy(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{AWSPhoneHomeCMKARN: testPhoneHomeCMKARN}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	var policy *map[string]any
	for i := range role.Policies {
		if role.Policies[i].PolicyName == "PhoneHomeSecretPolicy" {
			doc := role.Policies[i].PolicyDocument.(map[string]any)
			policy = &doc
		}
	}
	require.NotNil(t, policy, "expected a PhoneHomeSecretPolicy inline policy")

	statements := (*policy)["Statement"].([]map[string]any)
	require.Len(t, statements, 2, "expected a GetSecretValue and a kms:Decrypt statement")

	assert.Equal(t, testPhoneHomeSecretARN, statements[0]["Resource"])
	assert.Equal(t, []string{"secretsmanager:GetSecretValue"}, statements[0]["Action"])
	assert.Equal(t, testPhoneHomeCMKARN, statements[1]["Resource"])
	assert.Equal(t, []string{"kms:Decrypt", "kms:DescribeKey"}, statements[1]["Action"])
}

func TestGetRunnerPhoneHomeLambdaRole_OmitsKMSWhenUnset(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	for i := range role.Policies {
		if role.Policies[i].PolicyName != "PhoneHomeSecretPolicy" {
			continue
		}
		doc := role.Policies[i].PolicyDocument.(map[string]any)
		statements := doc["Statement"].([]map[string]any)
		assert.Len(t, statements, 1, "no CMK configured means no kms statement")

		return
	}
	t.Fatal("expected a PhoneHomeSecretPolicy inline policy")
}

func TestGetRunnerPhoneHomeProps_DoesNotEchoSecretARN(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{AWSPhoneHomeCMKARN: testPhoneHomeCMKARN}}
	inp := phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")

	props := tpl.getRunnerPhoneHomeProps(inp, nil)

	require.NotNil(t, props)
	for key, value := range props.Properties {
		str, ok := value.(string)
		if !ok {
			continue
		}
		assert.NotContains(t, str, testPhoneHomeSecretARN, "property %q echoes the secret ARN", key)
		assert.NotContains(t, str, "secretsmanager", "property %q echoes a secrets manager reference", key)
	}
}

func TestGetRunnerPhoneHomeLambda_TimeoutAndMemory(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}

	for _, tc := range []struct {
		name string
		inp  *stacks.TemplateInput
	}{
		{"with phone home auth", phoneHomeAuthTestInput("instabcdefghijklmnopqrstuv")},
		{"without phone home auth", phoneHomeTestInput("instabcdefghijklmnopqrstuv")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fn := tpl.getRunnerPhoneHomeLambda(tc.inp, tagBuilder{installID: tc.inp.Install.ID})

			require.NotNil(t, fn.Timeout, "unset means CloudFormation's 3s default, which is too short for the token fetch")
			require.NotNil(t, fn.MemorySize, "unset means CloudFormation's 128MB default, which is too little CPU to import boto3 in time")

			const scriptRetrySleepSeconds = 26.25
			assert.Greater(t, float64(*fn.Timeout), scriptRetrySleepSeconds,
				"timeout must clear the script's %.2fs of retry sleeps, or the ladder can never finish",
				scriptRetrySleepSeconds)

			assert.GreaterOrEqual(t, *fn.MemorySize, 512,
				"below ~512MB the boto3 import is slow enough to threaten the timeout again")
		})
	}
}

func TestValidatePhoneHomeScript(t *testing.T) {
	assert.NoError(t, validatePhoneHomeScript(strings.Repeat("x", lambdaInlineCodeLimit)))

	err := validatePhoneHomeScript(strings.Repeat("x", lambdaInlineCodeLimit+1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inline lambda source")

	err = validatePhoneHomeScript("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestGetRunnerPhoneHomeLambdaRole_DeterministicName(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	role := tpl.getRunnerPhoneHomeLambdaRole(inp, tagBuilder{installID: inp.Install.ID})

	require.NotNil(t, role.RoleName)
	assert.Equal(t, "instabcdefghijklmnopqrstuv-phone-home", *role.RoleName)
	assert.Equal(t, stacks.PhoneHomeRoleName(inp.Install.ID), *role.RoleName)
}

func TestPhoneHomeRoleName_WithinIAMLimit(t *testing.T) {
	name := stacks.PhoneHomeRoleName(strings.Repeat("i", 26))

	assert.LessOrEqual(t, len(name), 64,
		"IAM role names cannot exceed 64 characters")
	assert.Equal(t, strings.Repeat("i", 26)+"-phone-home", name)
}

func TestGetRunnerPhoneHomeProps_DoesNotEchoRoleName(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")

	props := tpl.getRunnerPhoneHomeProps(inp, nil)

	require.NotNil(t, props)
	for key, value := range props.Properties {
		if str, ok := value.(string); ok {
			assert.NotContains(t, str, "-phone-home",
				"property %q echoes the phone-home role name", key)
		}
	}
}

func TestGetRunnerPhoneHomeProps_NamedPolicyARNs(t *testing.T) {
	tpl := &Templates{cfg: &internal.Config{}}
	inp := phoneHomeTestInput("instabcdefghijklmnopqrstuv")
	inp.AppCfg.PermissionsConfig.NamedPolicies = []app.AppNamedIAMPolicyConfig{
		{
			Name:                    "grafana-lgtm-cloudwatch",
			CloudFormationStackName: "NamedPolicyGrafanaLgtmCloudwatch",
		},
	}

	props := tpl.getRunnerPhoneHomeProps(inp, nil)

	namedPolicyARNs, ok := props.Properties["named_policy_arns"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(
		t,
		cloudformation.Ref("NamedPolicyGrafanaLgtmCloudwatch"),
		namedPolicyARNs["grafana-lgtm-cloudwatch"],
	)
}
