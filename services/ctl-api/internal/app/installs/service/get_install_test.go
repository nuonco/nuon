package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallByID() {
	install := s.createTestInstall()
	policy := app.InstallManagementPolicyVersion{
		InstallID:         install.ID,
		Version:           1,
		EffectiveAt:       time.Now(),
		Connectivity:      app.InstallConnectivityDisconnected,
		ReleaseSelection:  app.InstallReleaseSelectionCustomer,
		CommandAuthority:  app.InstallAuthorityCustomer,
		ApprovalAuthority: app.InstallAuthorityCustomer,
		Telemetry:         app.InstallTelemetryManual,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&policy).Error)

	path := fmt.Sprintf("/v1/installs/%s", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), install.ID, resp.ID)
	assert.Equal(s.T(), install.Name, resp.Name)
	assert.Equal(s.T(), s.testApp.ID, resp.AppID)
	require.NotNil(s.T(), resp.ManagementPolicy)
	assert.Equal(s.T(), policy.ID, resp.ManagementPolicy.ID)
}

func (s *InstallsServiceTestSuite) TestGetInstallByName() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s", install.Name)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), install.ID, resp.ID)
}

func (s *InstallsServiceTestSuite) TestGetInstallNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000", nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallResolvesCloudPlatform() {
	for _, tc := range s.cloudPlatformResolutionTestCases() {
		s.Run(tc.name, func() {
			install := tc.setup()

			path := fmt.Sprintf("/v1/installs/%s", install.ID)
			rr := s.makeRequest(http.MethodGet, path, nil)
			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var resp app.Install
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
			assert.Equal(s.T(), tc.expectedCloudPlatform, resp.CloudPlatform)
			assert.Equal(s.T(), tc.expectedRunnerType, resp.RunnerType)
		})
	}
}
