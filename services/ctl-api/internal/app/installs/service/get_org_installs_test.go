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

func (s *InstallsServiceTestSuite) TestGetOrgInstallsEmpty() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsReturnsList() {
	install := s.createTestInstall()
	s.createTestInstall()
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

	rr := s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)
	for _, returnedInstall := range resp {
		if returnedInstall.ID == install.ID {
			require.NotNil(s.T(), returnedInstall.ManagementPolicy)
			assert.Equal(s.T(), policy.ID, returnedInstall.ManagementPolicy.ID)
			return
		}
	}
	s.T().Fatalf("install %s was not returned", install.ID)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsSearch() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs?q=%s", install.Name)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), install.ID, resp[0].ID)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsSearchByBranchName() {
	connected := s.createTestInstall()
	s.createTestInstall()

	branch := &app.AppBranch{
		AppID: s.testApp.ID,
		OrgID: s.testOrg.ID,
		Name:  "release-candidate",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{}).
		Where(app.Install{ID: connected.ID}).
		Update("app_branch_id", branch.ID).Error)

	path := fmt.Sprintf("/v1/installs?q=%s", branch.Name)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), connected.ID, resp[0].ID)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsFiltersByBranchNames() {
	onMain := s.createTestInstall()
	onRelease := s.createTestInstall()
	unconnected := s.createTestInstall()

	main := &app.AppBranch{AppID: s.testApp.ID, OrgID: s.testOrg.ID, Name: "main"}
	release := &app.AppBranch{AppID: s.testApp.ID, OrgID: s.testOrg.ID, Name: "release"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(main).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(release).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{}).Where(app.Install{ID: onMain.ID}).
		Update("app_branch_id", main.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{}).Where(app.Install{ID: onRelease.ID}).
		Update("app_branch_id", release.ID).Error)

	var resp []app.Install

	rr := s.makeRequest(http.MethodGet, "/v1/installs?branches=main", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), onMain.ID, resp[0].ID)

	rr = s.makeRequest(http.MethodGet, "/v1/installs?branches=main,release", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)

	rr = s.makeRequest(http.MethodGet, "/v1/installs?branches=__none__", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), unconnected.ID, resp[0].ID)

	rr = s.makeRequest(http.MethodGet, "/v1/installs?branches=main,__none__", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)

	rr = s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 3)
}

func (s *InstallsServiceTestSuite) TestGetOrgInstallsResolvesCloudPlatform() {
	for _, tc := range s.cloudPlatformResolutionTestCases() {
		s.Run(tc.name, func() {
			install := tc.setup()

			rr := s.makeRequest(http.MethodGet, "/v1/installs", nil)
			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var resp []app.Install
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

			found := findInstallByID(resp, install.ID)
			require.NotNil(s.T(), found, "install %s should be present in org installs list", install.ID)
			assert.Equal(s.T(), tc.expectedCloudPlatform, found.CloudPlatform)
			assert.Equal(s.T(), tc.expectedRunnerType, found.RunnerType)
		})
	}
}
