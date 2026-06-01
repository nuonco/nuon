package cloudformation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func TestGetRunnerParameters_InstanceType(t *testing.T) {
	tpl := &Templates{}

	t.Run("defaults to t3a.medium when unset", func(t *testing.T) {
		inp := &stacks.TemplateInput{Settings: &app.RunnerGroupSettings{}}
		params := tpl.getRunnerParameters(inp)

		p := params["RunnerInstanceType"]
		assert.Equal(t, app.DefaultAWSInstanceType, p.Default)
		assert.Contains(t, p.AllowedValues, interface{}(app.DefaultAWSInstanceType))
	})

	t.Run("uses configured value already in the allowed list", func(t *testing.T) {
		inp := &stacks.TemplateInput{Settings: &app.RunnerGroupSettings{AWSInstanceType: "t3.large"}}
		params := tpl.getRunnerParameters(inp)

		p := params["RunnerInstanceType"]
		assert.Equal(t, "t3.large", p.Default)
		assert.Contains(t, p.AllowedValues, interface{}("t3.large"))
		assert.Len(t, p.AllowedValues, 6, "value already present should not be appended")
	})

	t.Run("auto-adds a custom value to allowed values", func(t *testing.T) {
		inp := &stacks.TemplateInput{Settings: &app.RunnerGroupSettings{AWSInstanceType: "m5.xlarge"}}
		params := tpl.getRunnerParameters(inp)

		p := params["RunnerInstanceType"]
		assert.Equal(t, "m5.xlarge", p.Default)
		require.Contains(t, p.AllowedValues, interface{}("m5.xlarge"))
		assert.Len(t, p.AllowedValues, 7, "custom value should be appended once")
	})
}
