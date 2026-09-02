package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) seedRolesOnNewerConfig(installID string) *app.AppConfig {
	newerCfg := s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), s.testApp.ID)
	perm := &app.AppPermissionsConfig{
		AppID:       s.testApp.ID,
		AppConfigID: newerCfg.ID,
		Roles: []app.AppAWSIAMRoleConfig{
			{
				AppConfigID:   newerCfg.ID,
				CloudPlatform: string(app.CloudPlatformAzure),
				Type:          app.AWSIAMRoleTypeRunnerProvision,
				Name:          "{{.nuon.install.id}}-provision",
				DisplayName:   "provision role",
			},
			{
				AppConfigID:   newerCfg.ID,
				CloudPlatform: string(app.CloudPlatformAzure),
				Type:          app.AWSIAMRoleTypeRunnerMaintenance,
				Name:          "{{.nuon.install.id}}-maintenance",
				DisplayName:   "maintenance role",
			},
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(perm).Error)

	for i, role := range perm.Roles {
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.InstallRoles{
			InstallID:       installID,
			AppRoleConfigID: role.ID,
			Enabled:         true,
			Provisioned:     true,
			RoleID:          fmt.Sprintf("00000000-0000-0000-0000-00000000000%d", i),
		}).Error)
	}
	return newerCfg
}

// Live install roles follow the app's newest permissions config while the install
// stays pinned to the config it was created against, so the lookup must not be
// scoped to install.app_config_id.
func (s *InstallsServiceTestSuite) TestGetLatestInstallRolesForInstallPinnedToOlderConfig() {
	install := s.createTestInstall()
	newerCfg := s.seedRolesOnNewerConfig(install.ID)
	require.NotEqual(s.T(), newerCfg.ID, install.AppConfigID)

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/installs/%s/roles/latest", install.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var roles []app.InstallRoles
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &roles))
	require.Len(s.T(), roles, 2)
	for _, r := range roles {
		assert.True(s.T(), r.Provisioned)
		assert.NotEmpty(s.T(), r.RoleID)
		assert.Equal(s.T(), newerCfg.ID, r.AppRoleConfig.AppConfigID)
	}
}

func (s *InstallsServiceTestSuite) TestGetLatestInstallRolesSearch() {
	install := s.createTestInstall()
	s.seedRolesOnNewerConfig(install.ID)

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/installs/%s/roles/latest?q=maintenance", install.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var roles []app.InstallRoles
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &roles))
	require.Len(s.T(), roles, 1)
	assert.Equal(s.T(), "maintenance role", roles[0].AppRoleConfig.DisplayName)
}

func (s *InstallsServiceTestSuite) TestGetLatestInstallRolesNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/inl_nonexistent_0000000000/roles/latest", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
