package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestInstallTelemetrySettings() {
	install := s.createTestInstall()
	runnerGroup := &app.RunnerGroup{
		OrgID:     install.OrgID,
		OwnerID:   install.ID,
		OwnerType: "installs",
		Type:      app.RunnerGroupTypeInstall,
		Platform:  app.AppRunnerTypeAWS,
		Settings: app.RunnerGroupSettings{
			ContainerImageURL:      "registry.example.com/runner",
			ContainerImageTag:      "latest",
			RunnerAPIURL:           "https://runner.example.com",
			VendorTelemetryEnabled: true,
			Metadata:               pgtype.Hstore{},
			AWSMaxInstanceLifetime: 604800,
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(runnerGroup).Error)

	path := fmt.Sprintf("/v1/installs/%s/telemetry", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var settings InstallTelemetrySettings
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &settings))
	require.True(s.T(), settings.Enabled)

	rr = s.makeRequest(http.MethodPatch, path, UpdateInstallTelemetryRequest{Enabled: boolPtr(false)})
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &settings))
	require.False(s.T(), settings.Enabled)

	var persisted app.RunnerGroupSettings
	require.NoError(s.T(), s.deps.DB.Where(app.RunnerGroupSettings{RunnerGroupID: runnerGroup.ID}).First(&persisted).Error)
	require.False(s.T(), persisted.VendorTelemetryEnabled)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallTelemetrySettingsRequiresEnabled() {
	install := s.createTestInstall()
	rr := s.makeRequest(http.MethodPatch, fmt.Sprintf("/v1/installs/%s/telemetry", install.ID), map[string]any{})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestInstallTelemetrySettingsNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/inst-not-found/telemetry", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func boolPtr(value bool) *bool {
	return &value
}
