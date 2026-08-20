package service

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	s.createTestInstall()
	s.createTestInstall()

	rr := s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)
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

func (s *InstallsServiceTestSuite) TestGetOrgInstallsFiltersByBranchStatus() {
	connected := s.createTestInstall()
	unconnected := s.createTestInstall()

	branch := &app.AppBranch{
		AppID: s.testApp.ID,
		OrgID: s.testOrg.ID,
		Name:  "main",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{}).
		Where(app.Install{ID: connected.ID}).
		Update("app_branch_id", branch.ID).Error)

	var resp []app.Install

	rr := s.makeRequest(http.MethodGet, "/v1/installs?branch_status=assigned", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), connected.ID, resp[0].ID)

	rr = s.makeRequest(http.MethodGet, "/v1/installs?branch_status=none", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), unconnected.ID, resp[0].ID)

	rr = s.makeRequest(http.MethodGet, "/v1/installs", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)
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
