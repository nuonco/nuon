package service

import (
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestForgetInstallComponentStillInConfig() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	path := fmt.Sprintf("/v1/installs/%s/components/%s/forget", install.ID, helmComp.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestForgetInstallComponentStillInConfigUnchangedAcrossVersion() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	newCfg := s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(newCfg).
		Update("component_ids", pq.StringArray{helmComp.ID}).Error)

	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).
		Update("app_config_id", newCfg.ID).Error)

	path := fmt.Sprintf("/v1/installs/%s/components/%s/forget", install.ID, helmComp.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestForgetInstallComponentAfterAppComponentDeleted() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Delete(&app.Component{ID: helmComp.ID}).Error)

	path := fmt.Sprintf("/v1/installs/%s/components/%s/forget", install.ID, helmComp.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *InstallsServiceTestSuite) TestForgetInstallComponentNotFound() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components/cmp_nonexistent_00000000/forget", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
