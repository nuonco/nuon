package activities

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const mockOutputTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Resources:
  Bucket:
    Type: Custom::Bucket
Outputs:
  BucketArn:
    Value: !Ref Bucket
  BucketName:
    Value: !Ref Bucket
`

func TestRenderAWSStackTemplate_StackParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/runner" {
			_, _ = w.Write([]byte("Parameters:\n  InstanceType: {Type: String, Default: t3.small}\n"))
			return
		}
		_, _ = w.Write([]byte("Parameters:\n  VpcCIDR: {Type: String, Default: 10.0.0.0/16}\n  OptionalValue: {Type: String, Default: unused}\n"))
	}))
	t.Cleanup(server.Close)

	tests := map[string]struct {
		connected  bool
		parameters map[string]string
		want       map[string]string
		wantErr    string
	}{
		"render values from shared state without changing defaults": {
			connected: true,
			parameters: map[string]string{
				"RunnerInstanceType": "{{ .nuon.inputs.inputs.instance_type }}",
				"VpcCIDR":            "10.42.0.0/16",
				"OptionalValue":      "",
			},
			want: map[string]string{"RunnerInstanceType": "m6i.large", "VpcCIDR": "10.42.0.0/16", "OptionalValue": ""},
		},
		"overrides without a connection are rejected": {
			parameters: map[string]string{"RunnerInstanceType": "m6i.large"},
			wantErr:    "require an AWS cloud connection",
		},
		"undeclared parameter names are rejected": {
			connected:  true,
			parameters: map[string]string{"NotAStackParameter": "value"},
			wantErr:    `parameter "NotAStackParameter" is not declared`,
		},
		"missing template references are rejected": {
			connected:  true,
			parameters: map[string]string{"RunnerInstanceType": "{{ .nuon.inputs.inputs.missing }}"},
			wantErr:    "unable to render root stack parameters",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := stacks.TemplateInput{
				Install: &app.Install{
					ID: "test-install-id", AppID: "test-app-id", OrgID: "test-org-id",
					InstallConfig: &app.InstallConfig{AWSStackParameters: tt.parameters},
				},
				InstallState: &state.State{Inputs: &state.InputsState{Inputs: map[string]string{"instance_type": "m6i.large"}}},
				AppCfg: &app.AppConfig{StackConfig: app.AppStackConfig{
					VPCNestedTemplateURL: server.URL + "/vpc", RunnerNestedTemplateURL: server.URL + "/runner",
				}},
				Settings:                   &app.RunnerGroupSettings{},
				CloudFormationStackVersion: &app.InstallStackVersion{PhoneHomeURL: server.URL + "/phone-home"},
				PhonehomeScript:            "print('phone home')",
			}
			if tt.connected {
				connectionID := "test-cloud-connection"
				inp.Install.CloudConnectionID = &connectionID
			}
			a := &Activities{cfTemplates: cloudformation.NewTemplates(cloudformation.Params{Cfg: &internal.Config{UseLocalRunners: true}})}
			res, err := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, res.AWSStackParameters)
			assert.Equal(t, "{{ .nuon.inputs.inputs.instance_type }}", tt.parameters["RunnerInstanceType"])
			var tmpl struct {
				Parameters map[string]struct{ Default any }
			}
			require.NoError(t, json.Unmarshal(res.RAWJson, &tmpl))
			assert.Equal(t, "t3.small", tmpl.Parameters["RunnerInstanceType"].Default)
			assert.Equal(t, "10.0.0.0/16", tmpl.Parameters["VpcCIDR"].Default)
		})
	}
}

func TestRenderAWSStackTemplate_NestedParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templates := map[string]string{
			"/vpc":      "Parameters:\n  VpcCIDR: {Type: String, Default: 10.0.0.0/16}\nOutputs:\n  VPC: {Value: vpc-default}\n",
			"/runner":   "Parameters:\n  InstanceType: {Type: String, Default: t3.small}\n  RootVolumeSize: {Type: Number, Default: 30}\n  ExtraOption: {Type: String, Default: unused}\n",
			"/database": "Parameters:\n  InstanceClass: {Type: String}\n  BackupRetentionDays: {Type: String, Default: '1'}\n  OptionalValue: {Type: String, Default: unused}\n  VPC: {Type: String}\n",
		}
		_, _ = w.Write([]byte(templates[r.URL.Path]))
	}))
	t.Cleanup(server.Close)
	const values = `
[cloudformation.root.parameters]
RunnerRootVolumeSize = "60"
[cloudformation.nested_stacks.VPC.parameters]
VpcCIDR = "10.42.0.0/16"
[cloudformation.nested_stacks.RunnerAutoScalingGroup.parameters]
InstanceType = "m7i.large"
ExtraOption = "direct-value"
[cloudformation.nested_stacks.database.parameters]
InstanceClass = "db.r7g.large"
OptionalValue = ""
VPC = "vpc-external"
`
	for name, tt := range map[string]struct {
		values       string
		disconnected bool
		localRunner  bool
		customName   string
		wantErr      string
	}{
		"literals need no Nuon input declarations":      {values: values},
		"templates render without HTML escaping":        {values: strings.Replace(values, "direct-value", "{{ .nuon.inputs.inputs.option }}", 1)},
		"unknown stack is rejected":                     {values: "[cloudformation.nested_stacks.missing.parameters]\nValue = 'x'", wantErr: `nested stack "missing" does not exist`},
		"unknown nested parameter is rejected":          {values: "[cloudformation.nested_stacks.database.parameters]\nMissing = 'x'", wantErr: `parameter "Missing" is not declared`},
		"missing nested template reference is rejected": {values: "[cloudformation.nested_stacks.database.parameters]\nInstanceClass = '{{ .nuon.inputs.inputs.missing }}'", wantErr: "unable to render"},
		"nested-only overrides require a connection":    {values: values, disconnected: true, wantErr: "require an AWS cloud connection"},
		"runner target must exist":                      {values: values, localRunner: true, wantErr: `nested stack "RunnerAutoScalingGroup" does not exist`},
		"built-in names cannot name custom stacks":      {values: values, customName: "VPC", wantErr: "conflicts with a built-in stack"},
	} {
		t.Run(name, func(t *testing.T) {
			var installCfg config.Install
			require.NoError(t, toml.Unmarshal([]byte("cloud_connection = 'acme-aws'\n"+tt.values), &installCfg))
			require.NoError(t, installCfg.Validate())
			connectionID := "test-cloud-connection"
			inp := stacks.TemplateInput{
				Install: &app.Install{
					ID: "test-install", AppID: "test-app", OrgID: "test-org", CloudConnectionID: &connectionID,
					InstallConfig: &app.InstallConfig{AWSStackParameters: installCfg.CloudFormation.Root.Parameters, AWSNestedStackParameters: installCfg.CloudFormation.NestedStacks},
				},
				InstallState: &state.State{Inputs: &state.InputsState{Inputs: map[string]string{"option": "direct&value=\"quoted\""}}},
				AppCfg: &app.AppConfig{StackConfig: app.AppStackConfig{
					Type: app.StackTypeAWS, VPCNestedTemplateURL: server.URL + "/vpc", RunnerNestedTemplateURL: server.URL + "/runner",
					CustomNestedStacks: []config.CustomNestedStack{{Name: "database", TemplateURL: server.URL + "/database", Index: 20,
						Parameters: map[string]string{"InstanceClass": "{{ .nuon.inputs.inputs.unavailable }}", "BackupRetentionDays": "7"}}},
				}},
				Runner: &app.Runner{ID: "test-runner"}, Settings: &app.RunnerGroupSettings{},
				CloudFormationStackVersion: &app.InstallStackVersion{PhoneHomeURL: server.URL + "/phone-home"},
				PhonehomeScript:            "print('phone home')",
			}
			if tt.disconnected {
				inp.Install.CloudConnectionID = nil
			}
			if tt.customName != "" {
				inp.AppCfg.StackConfig.CustomNestedStacks[0].Name = tt.customName
			}
			original, err := json.Marshal(installCfg.CloudFormation)
			require.NoError(t, err)
			if tt.wantErr == "" {
				app.ApplyInstallStackOverrides(inp.Install, &inp.AppCfg.StackConfig)
				data, err := inp.InstallState.AsMap()
				require.NoError(t, err)
				require.NoError(t, config.RenderCustomNestedStackParameters(inp.AppCfg.StackConfig.CustomNestedStacks, data))
			}
			a := &Activities{cfTemplates: cloudformation.NewTemplates(cloudformation.Params{Cfg: &internal.Config{UseLocalRunners: tt.localRunner}})}
			res, err := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			var tmpl struct {
				Parameters map[string]struct{ Default any }
				Resources  map[string]struct {
					Properties struct {
						Parameters  map[string]any
						TemplateURL any
					}
					DependsOn []string
				}
			}
			require.NoError(t, json.Unmarshal(res.RAWJson, &tmpl))
			assert.Equal(t, map[string]string{"RunnerRootVolumeSize": "60"}, res.AWSStackParameters)
			assert.Equal(t, "30", tmpl.Parameters["RunnerRootVolumeSize"].Default)
			assert.Equal(t, "10.42.0.0/16", tmpl.Resources["VPC"].Properties.Parameters["VpcCIDR"])
			assert.NotContains(t, tmpl.Parameters, "VpcCIDR")
			runner := tmpl.Resources["RunnerAutoScalingGroup"].Properties.Parameters
			assert.Equal(t, "m7i.large", runner["InstanceType"])
			assert.Equal(t, map[string]any{"Ref": "RunnerRootVolumeSize"}, runner["RootVolumeSize"])
			wantOption := "direct-value"
			if name == "templates render without HTML escaping" {
				wantOption = "direct&value=\"quoted\""
			}
			assert.Equal(t, wantOption, runner["ExtraOption"])
			database := tmpl.Resources["Database"]
			assert.Equal(t, map[string]any{"InstanceClass": "db.r7g.large", "BackupRetentionDays": "7", "OptionalValue": "", "VPC": "vpc-external"}, database.Properties.Parameters)
			assert.Equal(t, map[string]any{"Fn::Join": []any{"", []any{server.URL + "/database"}}}, database.Properties.TemplateURL)
			assert.Equal(t, []string{"VPC", "RunnerAutoScalingGroup"}, database.DependsOn)
			assert.Equal(t, 20, inp.AppCfg.StackConfig.CustomNestedStacks[0].Index)
			after, err := json.Marshal(installCfg.CloudFormation)
			require.NoError(t, err)
			assert.Equal(t, original, after)
			next, err := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
			require.NoError(t, err)
			assert.Equal(t, res.RAWJson, next.RAWJson)
			inp.Install.InstallConfig.AWSNestedStackParameters["database"].Parameters["InstanceClass"] = "db.r7g.xlarge"
			next, err = a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
			require.NoError(t, err)
			assert.NotEqual(t, res.Checksum, next.Checksum)
		})
	}
}

func TestRenderAWSStackTemplate_CustomStacksOutputMapAgreesWithSplit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockOutputTemplateYAML))
	}))
	defer server.Close()

	customStacks := []config.CustomNestedStack{
		{Name: "my-bucket", TemplateURL: server.URL, Index: 0},
	}

	inp := stacks.TemplateInput{
		Install: &app.Install{ID: "test-install-id", AppID: "test-app-id", OrgID: "test-org-id"},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{CustomNestedStacks: customStacks},
		},
		Settings:         &app.RunnerGroupSettings{},
		CustomStacksOnly: true,
	}

	a := &Activities{cfTemplates: cloudformation.NewTemplates(cloudformation.Params{Cfg: &internal.Config{}})}

	res, err := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
	require.NoError(t, err)

	tmpl, _, err := a.cfTemplates.Template(&inp)
	require.NoError(t, err)

	flatOutputs := make(map[string]string, len(tmpl.Outputs))
	for name := range tmpl.Outputs {
		flatOutputs[name] = name
	}
	want := map[string]map[string]string{}
	for stackName, stackResult := range cloudformation.SplitCustomStacksOnlyOutputs(flatOutputs, []string{"my-bucket"}) {
		want[stackName] = stackResult["outputs"]
	}

	assert.Equal(t, want, res.CustomStacksOutputMap)
	assert.Equal(t, map[string]string{
		"BucketArn":  "MyBucketBucketArn",
		"BucketName": "MyBucketBucketName",
	}, res.CustomStacksOutputMap["my-bucket"])
}

func TestRenderAWSStackTemplate_CustomStacksOutputMapEmptyWhenNotCustomStacksOnly(t *testing.T) {
	a := &Activities{cfTemplates: cloudformation.NewTemplates(cloudformation.Params{Cfg: &internal.Config{}})}

	inp := stacks.TemplateInput{
		Install:          &app.Install{ID: "test-install-id", AppID: "test-app-id", OrgID: "test-org-id"},
		AppCfg:           &app.AppConfig{},
		Settings:         &app.RunnerGroupSettings{},
		CustomStacksOnly: false,
	}

	res, _ := a.RenderAWSStackTemplate(context.Background(), &RenderAWSStackTemplateRequest{Input: inp})
	assert.Nil(t, res.CustomStacksOutputMap)
}
