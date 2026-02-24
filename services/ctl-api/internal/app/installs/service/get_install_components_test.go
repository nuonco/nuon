package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallComponentsReturnsList() {
	install := s.createTestInstall()

	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	tfComp := s.getSeededComponent(app.ComponentTypeTerraformModule)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, tfComp.ID)

	path := fmt.Sprintf("/v1/installs/%s/components", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.InstallComponent
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 2)
}

func (s *InstallsServiceTestSuite) TestGetInstallComponentsFilterByType() {
	install := s.createTestInstall()

	helmComp := s.getSeededComponent(app.ComponentTypeHelmChart)
	tfComp := s.getSeededComponent(app.ComponentTypeTerraformModule)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, helmComp.ID)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, tfComp.ID)

	path := fmt.Sprintf("/v1/installs/%s/components?types=helm_chart", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.InstallComponent
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(s.T(), resp, 1)
}

func (s *InstallsServiceTestSuite) TestGetInstallComponentsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.InstallComponent
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}
