package cloudformation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation/ec2"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func TestGetAWSTemplate_RunnerResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockVPCTemplateYAML))
	}))
	defer server.Close()

	newInput := func() *stacks.TemplateInput {
		return &stacks.TemplateInput{
			Install: &app.Install{
				ID:    "inl123",
				AppID: "app123",
				OrgID: "org123",
			},
			CloudFormationStackVersion: &app.InstallStackVersion{PhoneHomeURL: server.URL + "/phone-home"},
			AppCfg: &app.AppConfig{
				StackConfig: app.AppStackConfig{
					VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
					RunnerNestedTemplateURL: server.URL + "/runner.yaml",
				},
			},
			Runner:          &app.Runner{ID: "run123"},
			Settings:        &app.RunnerGroupSettings{},
			PhonehomeScript: "print('phone home')",
		}
	}

	// The sandbox terraform looks the runner security group up by tag to grant it
	// cluster access, so it has to exist even when no runner instance does.
	t.Run("runner security group is created with local runners", func(t *testing.T) {
		tpl := &Templates{cfg: &internal.Config{UseLocalRunners: true}}

		tmpl, err := tpl.getAWSTemplate(newInput())
		require.NoError(t, err)

		require.Contains(t, tmpl.Resources, "RunnerSecurityGroup")
		sg, ok := tmpl.Resources["RunnerSecurityGroup"].(*ec2.SecurityGroup)
		require.True(t, ok)
		assert.Contains(t, sg.Tags, tags.Tag{Key: "network.nuon.co/domain", Value: "runner"})

		assert.NotContains(t, tmpl.Resources, "RunnerAutoScalingGroup")
		assert.NotContains(t, tmpl.Resources, "RunnerCloudWatchLogGroup")
	})

	t.Run("runner asg and logs are created with cloud runners", func(t *testing.T) {
		tpl := &Templates{cfg: &internal.Config{UseLocalRunners: false}}

		tmpl, err := tpl.getAWSTemplate(newInput())
		require.NoError(t, err)

		assert.Contains(t, tmpl.Resources, "RunnerSecurityGroup")
		assert.Contains(t, tmpl.Resources, "RunnerAutoScalingGroup")
		assert.Contains(t, tmpl.Resources, "RunnerCloudWatchLogGroup")
		assert.Contains(t, tmpl.Resources, "RunnerCloudWatchLogStream")
		assert.Contains(t, tmpl.Resources, "RunnerCloudWatchLogPolicy")
	})
}

func TestGetAWSTemplate_TelemetryIngress(t *testing.T) {
	const telemetryRunner = `
Parameters:
  EnableTelemetryIngress: {Type: String, Default: "false"}
  VpcId: {Type: String, Default: ""}
  TelemetrySourcePrefixListId: {Type: String, Default: ""}
Outputs:
  TelemetryEndpoint: {Value: ""}
`
	const vpcOutputs = `
Outputs:
  VPC: {Value: vpc-example}
  VpcIpv4PrefixListId: {Value: pl-example}
`
	for _, tc := range []struct {
		name          string
		runner        string
		vpc           string
		local         bool
		wantSupported bool
	}{
		{"managed VPC", telemetryRunner, mockVPCTemplateYAML + vpcOutputs, false, true},
		{"BYO VPC without CIDR parameter", telemetryRunner, mockBYOVPCTemplateYAML + vpcOutputs, false, true},
		{"older runner template", mockRunnerASGTemplateYAML, mockVPCTemplateYAML + vpcOutputs, false, false},
		{"missing nested output", strings.ReplaceAll(telemetryRunner, "TelemetryEndpoint", "OtherEndpoint"), mockVPCTemplateYAML + vpcOutputs, false, false},
		{"missing network parameter", strings.ReplaceAll(telemetryRunner, "VpcId", "OtherVpcId"), mockVPCTemplateYAML + vpcOutputs, false, false},
		{"older CIDR-based runner", strings.ReplaceAll(telemetryRunner, "TelemetrySourcePrefixListId", "VpcCIDR"), mockVPCTemplateYAML + vpcOutputs, false, false},
		{"older managed VPC", telemetryRunner, mockVPCTemplateYAML, false, false},
		{"older BYO VPC", telemetryRunner, mockBYOVPCTemplateYAML, false, false},
		{"local runner", telemetryRunner, mockVPCTemplateYAML + vpcOutputs, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/vpc.yaml" {
					_, _ = w.Write([]byte(tc.vpc))
					return
				}
				_, _ = w.Write([]byte(tc.runner))
			}))
			defer server.Close()

			inp := &stacks.TemplateInput{
				Install:                    &app.Install{ID: "inl123", AppID: "app123", OrgID: "org123"},
				Runner:                     &app.Runner{ID: "run123"},
				Settings:                   &app.RunnerGroupSettings{},
				CloudFormationStackVersion: &app.InstallStackVersion{PhoneHomeURL: server.URL + "/phone-home"},
				PhonehomeScript:            "print('phone home')",
				AppCfg: &app.AppConfig{StackConfig: app.AppStackConfig{
					VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
					RunnerNestedTemplateURL: server.URL + "/runner.yaml",
				}},
			}
			tpl := &Templates{cfg: &internal.Config{UseLocalRunners: tc.local}}
			tmpl, err := tpl.getAWSTemplate(inp)
			require.NoError(t, err)
			raw, err := tmpl.JSON()
			require.NoError(t, err)
			var rendered map[string]any
			require.NoError(t, json.Unmarshal(raw, &rendered))
			resources := rendered["Resources"].(map[string]any)
			props := resources["PhoneHomeProps"].(map[string]any)["Properties"].(map[string]any)
			output := rendered["Outputs"].(map[string]any)["TelemetryEndpoint"].(map[string]any)

			if !tc.wantSupported {
				assert.NotContains(t, tmpl.Parameters, "EnableTelemetryIngress")
				assert.NotContains(t, tmpl.Conditions, "TelemetryIngressEnabled")
				assert.Equal(t, "", props["telemetry_endpoint"])
				assert.Equal(t, "", output["Value"])
				assert.NotContains(t, string(raw), "Outputs.TelemetryEndpoint")
				if !tc.local {
					params := resources["RunnerAutoScalingGroup"].(map[string]any)["Properties"].(map[string]any)["Parameters"].(map[string]any)
					assert.NotContains(t, params, "EnableTelemetryIngress")
					assert.NotContains(t, params, "VpcId")
					assert.NotContains(t, params, "VpcCIDR")
					assert.NotContains(t, params, "TelemetrySourcePrefixListId")
				}
				return
			}

			assert.NotContains(t, tmpl.Parameters, "VpcId")
			assert.NotContains(t, tmpl.Parameters, "TelemetrySourcePrefixListId")
			parameter := tmpl.Parameters["EnableTelemetryIngress"]
			assert.Equal(t, "false", parameter.Default)
			assert.Equal(t, []any{"true", "false"}, parameter.AllowedValues)
			condition := rendered["Conditions"].(map[string]any)["TelemetryIngressEnabled"]
			assert.Equal(t, map[string]any{"Fn::Equals": []any{map[string]any{"Ref": "EnableTelemetryIngress"}, "true"}}, condition)
			params := resources["RunnerAutoScalingGroup"].(map[string]any)["Properties"].(map[string]any)["Parameters"].(map[string]any)
			assert.Equal(t, map[string]any{"Ref": "EnableTelemetryIngress"}, params["EnableTelemetryIngress"])
			assert.Equal(t, map[string]any{"Fn::GetAtt": []any{"VPC", "Outputs.VpcIpv4PrefixListId"}}, params["TelemetrySourcePrefixListId"])
			assert.NotContains(t, params, "VpcCIDR")
			assert.Equal(t, map[string]any{"Fn::GetAtt": []any{"VPC", "Outputs.VPC"}}, params["VpcId"])
			wantEndpoint := map[string]any{"Fn::If": []any{
				"TelemetryIngressEnabled",
				map[string]any{"Fn::GetAtt": []any{"RunnerAutoScalingGroup", "Outputs.TelemetryEndpoint"}},
				"",
			}}
			assert.Equal(t, wantEndpoint, props["telemetry_endpoint"])
			assert.Equal(t, wantEndpoint, output["Value"])
			groups := tmpl.Metadata["AWS::CloudFormation::Interface"].(map[string]any)["ParameterGroups"].([]map[string]any)
			assert.Contains(t, groups[0]["Parameters"], "EnableTelemetryIngress")
		})
	}
}
