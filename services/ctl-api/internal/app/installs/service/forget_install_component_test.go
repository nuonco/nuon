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

// TestForgetInstallComponentStillInConfigUnchangedAcrossVersion covers the case where a
// component is still in the install's app config but its ComponentConfigConnection is pinned to
// an earlier app config version (its config was unchanged across the version bump). The guard
// must consult AppConfig.ComponentIDs — the authoritative per-version list — rather than the
// version-pinned ComponentConfigConnections, otherwise the forget is wrongly allowed.
func (s *InstallsServiceTestSuite) TestForgetInstallComponentStillInConfigUnchangedAcrossVersion() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	// New config version that still lists the component but has no ComponentConfigConnections of
	// its own; the component's CCC stays pinned to the original version.
	newCfg := s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(newCfg).
		Update("component_ids", pq.StringArray{helmComp.ID}).Error)

	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).
		Update("app_config_id", newCfg.ID).Error)

	path := fmt.Sprintf("/v1/installs/%s/components/%s/forget", install.ID, helmComp.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

// TestForgetInstallComponentAfterAppComponentDeleted covers forgetting an install component whose
// app component no longer resolves. AppConfig.ComponentIDs is an immutable snapshot that still
// lists the component, so the guard must intersect with components that still exist — otherwise the
// orphaned install component could never be forgotten. A soft delete leaves the install component
// intact (a hard delete would cascade it away, leaving nothing to forget).
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
