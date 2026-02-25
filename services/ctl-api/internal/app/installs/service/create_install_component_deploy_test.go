package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestCreateComponentDeploySuccess() {
	install := s.createTestInstall()

	// Get component and config connection from the seeded app config.
	ccc := s.testAppConfig.ComponentConfigConnections[0]

	// Pre-seed the InstallComponent so the handler skips its auto-create branch
	// (which triggers a duplicate-table SQL error with the installs view).
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, ccc.ComponentID)

	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

	body := CreateInstallDeployRequest{
		BuildID: build.ID,
	}

	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, ccc.ComponentID)
	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var deploy app.InstallDeploy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &deploy))
	assert.NotEmpty(s.T(), deploy.ID)
	assert.Equal(s.T(), build.ID, deploy.ComponentBuildID)

	// Workflow should have been created.
	workflowID := rr.Header().Get(app.HeaderInstallWorkflowID)
	assert.NotEmpty(s.T(), workflowID)

	// Verify deploy persisted in DB.
	var dbDeploy app.InstallDeploy
	require.NoError(s.T(), s.deps.DB.Where("id = ?", deploy.ID).First(&dbDeploy).Error)
	assert.Equal(s.T(), build.ID, dbDeploy.ComponentBuildID)

	// Verify workflow persisted in DB.
	var dbWorkflow app.Workflow
	require.NoError(s.T(), s.deps.DB.Where("id = ?", workflowID).First(&dbWorkflow).Error)
	assert.Equal(s.T(), install.ID, dbWorkflow.OwnerID)
}

func (s *InstallsServiceTestSuite) TestCreateComponentDeployNotFound() {
	ccc := s.testAppConfig.ComponentConfigConnections[0]

	body := CreateInstallDeployRequest{
		BuildID: "fake-build",
	}

	path := fmt.Sprintf("/v1/installs/nonexistent/components/%s/deploys", ccc.ComponentID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateComponentDeployInvalidBuild() {
	install := s.createTestInstall()
	ccc := s.testAppConfig.ComponentConfigConnections[0]

	body := CreateInstallDeployRequest{
		BuildID: "nonexistent-build",
	}

	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, ccc.ComponentID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
