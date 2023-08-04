package helm

import (
	_ "embed"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultInstallValues(t *testing.T) {
	vals := NewDefaultInstallValues()
	assert.False(t, vals.Server.Enabled)
	assert.False(t, vals.UI.Service.Enabled)
	assert.True(t, vals.Runner.Enabled)

	// make sure the helm values have the right things disabled
	var helmVals map[string]interface{}
	err := mapstructure.Decode(vals, &helmVals)
	assert.Nil(t, err)

	// make sure the output server value is disabled
	serverVals := helmVals["server"].(map[string]interface{})
	enabled, ok := serverVals["enabled"]
	assert.True(t, ok)
	assert.False(t, enabled.(bool))

	// make sure the output server value is disabled
	uiVals := helmVals["ui"].(map[string]interface{})
	uiServiceValues, ok := uiVals["service"]
	assert.True(t, ok)
	enabled, ok = uiServiceValues.(map[string]interface{})["enabled"]
	assert.True(t, ok)
	assert.False(t, enabled.(bool))

	// assert that image is set
	assert.NotEmpty(t, vals.Runner.Image.Repository)
	assert.NotEmpty(t, vals.Runner.Image.Tag)
}

func TestNewDefaultOrgServerValues(t *testing.T) {
	vals := NewDefaultOrgServerValues()
	assert.True(t, vals.Server.Enabled)
	assert.False(t, vals.Runner.Enabled)
	assert.True(t, vals.Server.Enabled)
	assert.True(t, vals.UI.Service.Enabled)
	assert.False(t, vals.Bootstrap.ServiceAccount.Create)
	assert.False(t, vals.Runner.Enabled)

	// assert that image is set
	assert.NotEmpty(t, vals.Server.Image.Repository)
	assert.NotEmpty(t, vals.Server.Image.Tag)
}

func TestNewDefaultOrgRunnerValues(t *testing.T) {
	vals := NewDefaultOrgRunnerValues()

	assert.True(t, vals.Runner.Enabled)
	assert.Equal(t, 1, vals.Runner.Replicas)
	assert.True(t, vals.Runner.Server.TLS)
	assert.True(t, vals.Runner.Server.TLSSkipVerify)
	assert.True(t, vals.Runner.Odr.ServiceAccount.Create)

	// resources
	assert.Equal(t, "1Gi", vals.Runner.Storage.Size)
	assert.Equal(t, "256Mi", vals.Runner.Resources.Requests.Memory)
	assert.Equal(t, "250m", vals.Runner.Resources.Requests.CPU)

	// assert that image is set
	assert.NotEmpty(t, vals.Runner.Image.Repository)
	assert.NotEmpty(t, vals.Runner.Image.Tag)
}
