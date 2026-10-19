package cloudformation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func runnerTemplateInput(t *testing.T, configuredInstanceType string, template string) *stacks.TemplateInput {
	t.Helper()

	inp := &stacks.TemplateInput{
		Settings:                     &app.RunnerGroupSettings{},
		AppCfg:                       &app.AppConfig{},
		ConfiguredRunnerInstanceType: configuredInstanceType,
	}
	if template == "" {
		return inp
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(template))
	}))
	t.Cleanup(srv.Close)

	inp.AppCfg.StackConfig.RunnerNestedTemplateURL = srv.URL + "/stack.yaml"
	return inp
}

func TestGetRunnerParameters_InstanceType(t *testing.T) {
	tpl := &Templates{}

	t.Run("falls back to the platform default when nothing declares one", func(t *testing.T) {
		inp := runnerTemplateInput(t, "", "")
		params := tpl.getRunnerParameters(inp)

		p := params["RunnerInstanceType"]
		assert.Equal(t, app.DefaultAWSInstanceType, p.Default)
		assert.Contains(t, p.AllowedValues, interface{}(app.DefaultAWSInstanceType))
	})

	t.Run("uses configured value already in the allowed list", func(t *testing.T) {
		inp := runnerTemplateInput(t, "t3.large", "")
		params := tpl.getRunnerParameters(inp)

		p := params["RunnerInstanceType"]
		assert.Equal(t, "t3.large", p.Default)
		assert.Contains(t, p.AllowedValues, interface{}("t3.large"))
		assert.Len(t, p.AllowedValues, 6, "value already present should not be appended")
	})

	t.Run("auto-adds a custom value to allowed values", func(t *testing.T) {
		inp := runnerTemplateInput(t, "m5.xlarge", "")
		params := tpl.getRunnerParameters(inp)

		p := params["RunnerInstanceType"]
		assert.Equal(t, "m5.xlarge", p.Default)
		require.Contains(t, p.AllowedValues, interface{}("m5.xlarge"))
		assert.Len(t, p.AllowedValues, 7, "custom value should be appended once")
	})
}

const runnerTemplateWithDefaults = `
Parameters:
  InstanceType:
    Type: String
    Description: EC2 instance type for the runner
    Default: m6i.large
  RootVolumeSize:
    Type: Number
    Description: Size of the root EBS volume in GB
    Default: 100
    MinValue: 8
    MaxValue: 200
`

func TestGetRunnerParameters_NestedTemplateDefaults(t *testing.T) {
	tpl := &Templates{}

	t.Run("uses the nested template defaults when the runner config declares none", func(t *testing.T) {
		inp := runnerTemplateInput(t, "", runnerTemplateWithDefaults)
		params := tpl.getRunnerParameters(inp)

		instanceType := params["RunnerInstanceType"]
		assert.Equal(t, "m6i.large", instanceType.Default)
		assert.Contains(t, instanceType.AllowedValues, interface{}("m6i.large"))

		rootVolume := params["RunnerRootVolumeSize"]
		assert.Equal(t, "100", rootVolume.Default)
		assert.Equal(t, 8.0, *rootVolume.MinValue)
		assert.Equal(t, 200.0, *rootVolume.MaxValue)
	})

	// The stack generators resolve the platform default into Settings.AWSInstanceType before
	// rendering, which used to mask the template's own default and pin every install to
	// t3.medium no matter what the runner template declared.
	t.Run("a resolved settings value does not mask the template default", func(t *testing.T) {
		inp := runnerTemplateInput(t, "", runnerTemplateWithDefaults)
		inp.Settings.AWSInstanceType = app.DefaultAWSInstanceType

		params := tpl.getRunnerParameters(inp)

		assert.Equal(t, "m6i.large", params["RunnerInstanceType"].Default)
	})

	t.Run("the runner config wins over the nested template default", func(t *testing.T) {
		inp := runnerTemplateInput(t, "t3.large", runnerTemplateWithDefaults)
		params := tpl.getRunnerParameters(inp)

		assert.Equal(t, "t3.large", params["RunnerInstanceType"].Default)
	})

	t.Run("falls back to platform defaults when the template declares none", func(t *testing.T) {
		inp := runnerTemplateInput(t, "", "Parameters:\n  SubnetId:\n    Type: String\n")
		params := tpl.getRunnerParameters(inp)

		assert.Equal(t, app.DefaultAWSInstanceType, params["RunnerInstanceType"].Default)

		rootVolume := params["RunnerRootVolumeSize"]
		assert.Equal(t, "30", rootVolume.Default)
		assert.Equal(t, 8.0, *rootVolume.MinValue)
		assert.Equal(t, 100.0, *rootVolume.MaxValue)
	})

	t.Run("falls back to platform defaults when the template cannot be fetched", func(t *testing.T) {
		inp := runnerTemplateInput(t, "", "")
		inp.AppCfg.StackConfig.RunnerNestedTemplateURL = "http://127.0.0.1:1/stack.yaml"

		params := tpl.getRunnerParameters(inp)

		assert.Equal(t, app.DefaultAWSInstanceType, params["RunnerInstanceType"].Default)
		assert.Equal(t, "30", params["RunnerRootVolumeSize"].Default)
	})

	t.Run("widens the volume bounds so the template default stays valid", func(t *testing.T) {
		inp := runnerTemplateInput(t, "", "Parameters:\n  RootVolumeSize:\n    Type: Number\n    Default: 250\n")
		params := tpl.getRunnerParameters(inp)

		rootVolume := params["RunnerRootVolumeSize"]
		assert.Equal(t, "250", rootVolume.Default)
		assert.Equal(t, 250.0, *rootVolume.MaxValue)
	})
}

var runnerTemplateWithImageParams = strings.Replace(mockRunnerASGTemplateYAML, "Resources:\n", `  RunnerAmiAutoUpdate:
    Type: String
    Default: "false"
  RunnerAmiId:
    Type: String
    Default: ""
Resources:
`, 1)

func renderRunnerASGTemplate(t *testing.T, runnerTemplate string, runnerCfg app.AppRunnerConfig) (string, map[string]any) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/vpc.yaml" {
			_, _ = w.Write([]byte(mockVPCTemplateYAML))
			return
		}
		_, _ = w.Write([]byte(runnerTemplate))
	}))
	t.Cleanup(server.Close)

	tpl := &Templates{cfg: &internal.Config{}}
	tmpl, err := tpl.getAWSTemplate(&stacks.TemplateInput{
		Install:                    &app.Install{ID: "inl123", AppID: "app123", OrgID: "org123"},
		Runner:                     &app.Runner{ID: "run123"},
		Settings:                   &app.RunnerGroupSettings{},
		CloudFormationStackVersion: &app.InstallStackVersion{PhoneHomeURL: "https://example.com/phone-home"},
		PhonehomeScript:            "print('phone home')",
		AppCfg: &app.AppConfig{
			RunnerConfig: runnerCfg,
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL:    server.URL + "/vpc.yaml",
				RunnerNestedTemplateURL: server.URL + "/runner.yaml",
			},
		},
	})
	require.NoError(t, err)

	raw, err := tmpl.JSON()
	require.NoError(t, err)
	var rendered map[string]any
	require.NoError(t, json.Unmarshal(raw, &rendered))
	params := rendered["Resources"].(map[string]any)["RunnerAutoScalingGroup"].(map[string]any)["Properties"].(map[string]any)["Parameters"].(map[string]any)
	return strings.ReplaceAll(string(raw), server.URL, "https://templates.example.com"), params
}

func TestGetRunnerASGNestedStack_ImageParameters(t *testing.T) {
	tests := map[string]struct {
		runnerTemplate string
		runnerCfg      app.AppRunnerConfig
		wantAutoUpdate any
		wantImageID    any
	}{
		"older template ignores both fields": {
			runnerTemplate: mockRunnerASGTemplateYAML,
			runnerCfg:      app.AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0abc123"},
		},
		"template missing one parameter ignores both fields": {
			runnerTemplate: strings.Replace(runnerTemplateWithImageParams, "RunnerAmiId:", "OtherAmiId:", 1),
			runnerCfg:      app.AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0abc123"},
		},
		"unset fields leave the template defaults": {
			runnerTemplate: runnerTemplateWithImageParams,
		},
		"auto-update only": {
			runnerTemplate: runnerTemplateWithImageParams,
			runnerCfg:      app.AppRunnerConfig{ImageAutoUpdate: true},
			wantAutoUpdate: "true",
		},
		"pinned image only": {
			runnerTemplate: runnerTemplateWithImageParams,
			runnerCfg:      app.AppRunnerConfig{ImageID: "ami-0abc123"},
			wantImageID:    "ami-0abc123",
		},
		"auto-update and pinned image": {
			runnerTemplate: runnerTemplateWithImageParams,
			runnerCfg:      app.AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0abc123"},
			wantAutoUpdate: "true",
			wantImageID:    "ami-0abc123",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, params := renderRunnerASGTemplate(t, tc.runnerTemplate, tc.runnerCfg)

			assert.Equal(t, tc.wantAutoUpdate, params["RunnerAmiAutoUpdate"])
			assert.Equal(t, tc.wantImageID, params["RunnerAmiId"])
		})
	}
}

func TestGetAWSTemplate_ImageParametersPreserveExistingOutput(t *testing.T) {
	baseline, _ := renderRunnerASGTemplate(t, mockRunnerASGTemplateYAML, app.AppRunnerConfig{})

	tests := map[string]struct {
		runnerTemplate string
		runnerCfg      app.AppRunnerConfig
	}{
		"older template with fields set": {
			runnerTemplate: mockRunnerASGTemplateYAML,
			runnerCfg:      app.AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0abc123"},
		},
		"newer template with fields unset": {
			runnerTemplate: runnerTemplateWithImageParams,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rendered, _ := renderRunnerASGTemplate(t, tc.runnerTemplate, tc.runnerCfg)

			assert.Equal(t, baseline, rendered)
		})
	}
}
