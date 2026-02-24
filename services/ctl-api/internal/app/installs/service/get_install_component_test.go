package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallComponentSuccess() {
	install := s.createTestInstall()
	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	ic := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)

	path := fmt.Sprintf("/v1/installs/%s/components/%s", install.ID, ic.ComponentID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.InstallComponent
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), ic.ID, resp.ID)
	assert.Equal(s.T(), helmComp.ID, resp.ComponentID)
}

func (s *InstallsServiceTestSuite) TestGetInstallComponentNotFound() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components/cmp_nonexistent_00000000", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
