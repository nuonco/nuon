package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallComponentDeploysNoDeploys() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, helmComp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallComponentDeploysComponentNotFound() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components/cmp_nonexistent_00000000/deploys", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallComponentLatestDeployNoDeploys() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys/latest", install.ID, helmComp.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
