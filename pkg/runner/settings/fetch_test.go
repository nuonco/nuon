package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type settingsClient struct {
	nuonrunner.Client
	response *models.AppRunnerGroupSettings
}

func (c settingsClient) GetSettings(context.Context) (*models.AppRunnerGroupSettings, error) {
	return c.response, nil
}

func TestFetchKeepsVendorResourceAttributesSeparate(t *testing.T) {
	response := &models.AppRunnerGroupSettings{
		LoggingLevel: "info",
		Metadata:     map[string]string{"install.id": "install-test"},
		VendorTelemetryResourceAttributes: map[string]string{
			"nuon.install.name": "production", "nuon.install.labels.tier": "enterprise",
		},
	}
	s := &Settings{apiClient: settingsClient{response: response}, Cfg: &runnerconfig.Config{RunnerID: "runner-test"}}
	require.NoError(t, s.fetch(context.Background()))
	require.Equal(t, response.VendorTelemetryResourceAttributes, s.VendorTelemetryResourceAttributes)
	require.NotContains(t, s.Metadata, "nuon.install.labels.tier")
	require.Equal(t, "runner-test", s.Metadata["runner.id"])
	response.VendorTelemetryResourceAttributes = nil
	require.NoError(t, s.fetch(context.Background()))
	require.Nil(t, s.VendorTelemetryResourceAttributes, "an older API must not retain the previous snapshot")
}
