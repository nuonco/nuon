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

func (s *InstallsServiceTestSuite) TestGetAppInstallsReturnsList() {
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

	path := fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 1)
	assert.Equal(s.T(), s.testApp.ID, resp[0].AppID)
	require.NotNil(s.T(), resp[0].ManagementPolicy)
	assert.Equal(s.T(), policy.ID, resp[0].ManagementPolicy.ID)
}

func (s *InstallsServiceTestSuite) TestGetAppInstallsFiltersByAppBranch() {
	connected := s.createTestInstall()
	s.createTestInstall()

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

	path := fmt.Sprintf("/v1/apps/%s/installs?app_branch_id=%s", s.testApp.ID, branch.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(s.T(), resp, 1)
	assert.Equal(s.T(), connected.ID, resp[0].ID)

	rr = s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)
}

func (s *InstallsServiceTestSuite) TestGetAppInstallsEmpty() {
	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())

	path := fmt.Sprintf("/v1/apps/%s/installs", otherApp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}

func (s *InstallsServiceTestSuite) TestGetAppInstallsResolvesCloudPlatform() {
	for _, tc := range s.cloudPlatformResolutionTestCases() {
		s.Run(tc.name, func() {
			install := tc.setup()

			path := fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID)
			rr := s.makeRequest(http.MethodGet, path, nil)
			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var resp []app.Install
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

			found := findInstallByID(resp, install.ID)
			require.NotNil(s.T(), found, "install %s should be present in app installs list", install.ID)
			assert.Equal(s.T(), tc.expectedCloudPlatform, found.CloudPlatform)
			assert.Equal(s.T(), tc.expectedRunnerType, found.RunnerType)
		})
	}
}

func (s *InstallsServiceTestSuite) TestGetAppInstallsSerializesAzureAccount() {
	install := s.createTestInstall()

	azureAccount := &app.AzureAccount{
		InstallID:                install.ID,
		Location:                 "eastus",
		SubscriptionID:           "sub-1234",
		SubscriptionTenantID:     "tenant-1234",
		ServicePrincipalAppID:    "sp-app-1234",
		ServicePrincipalPassword: "sp-secret",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(azureAccount).Error)

	path := fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))

	found := findInstallByID(resp, install.ID)
	require.NotNil(s.T(), found, "install %s should be present in app installs list", install.ID)
	require.NotNil(s.T(), found.AzureAccount, "azure_account should be serialized on the app installs endpoint")
	assert.Equal(s.T(), azureAccount.ID, found.AzureAccount.ID)
	assert.Equal(s.T(), "eastus", found.AzureAccount.Location)
	assert.Equal(s.T(), "sub-1234", found.AzureAccount.SubscriptionID)
}
