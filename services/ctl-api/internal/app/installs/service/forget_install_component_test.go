package service

import (
	"fmt"
	"net/http"

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

func (s *InstallsServiceTestSuite) TestForgetInstallComponentNotFound() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components/cmp_nonexistent_00000000/forget", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
