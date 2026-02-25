package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetComponentsDeploysEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/components/deploys", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var deploys []app.InstallDeploy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &deploys))
	assert.Empty(s.T(), deploys)
}

func (s *InstallsServiceTestSuite) TestGetComponentsDeploysReturnsList() {
	install := s.createTestInstall()

	// Seed the full deploy chain: component -> install component -> build -> deploy.
	ccc := s.testAppConfig.ComponentConfigConnections[0]
	installComp := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, ccc.ComponentID)
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
	s.deps.Seeder.CreateInstallDeploy(s.ctx, s.T(), installComp.ID, build.ID)

	path := fmt.Sprintf("/v1/installs/%s/components/deploys", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var deploys []app.InstallDeploy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &deploys))
	assert.Len(s.T(), deploys, 1)
	assert.Equal(s.T(), build.ID, deploys[0].ComponentBuildID)
}
