package cloudformation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
