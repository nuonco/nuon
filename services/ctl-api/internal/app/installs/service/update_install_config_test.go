package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

func (s *InstallsServiceTestSuite) TestUpdateInstallConfigSuccess() {
	install := s.createTestInstall()

	createBody := CreateInstallConfigRequest{
		CreateInstallConfigParams: helpers.CreateInstallConfigParams{
			ApprovalOption: app.InstallApprovalOptionApproveAll,
		},
	}
	createPath := fmt.Sprintf("/v1/installs/%s/configs", install.ID)
	createRR := s.makeRequest(http.MethodPost, createPath, createBody)
	require.Equal(s.T(), http.StatusCreated, createRR.Code)

	var created app.InstallConfig
	require.NoError(s.T(), json.Unmarshal(createRR.Body.Bytes(), &created))

	promptOpt := app.InstallApprovalOptionPrompt
	updateBody := UpdateInstallConfigRequest{
		ApprovalOption: &promptOpt,
	}
	updatePath := fmt.Sprintf("/v1/installs/%s/configs/%s", install.ID, created.ID)
	rr := s.makeRequest(http.MethodPatch, updatePath, updateBody)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var updated app.InstallConfig
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &updated))
	assert.Equal(s.T(), created.ID, updated.ID)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallConfigNotFound() {
	install := s.createTestInstall()

	promptOpt := app.InstallApprovalOptionPrompt
	body := UpdateInstallConfigRequest{
		ApprovalOption: &promptOpt,
	}

	path := fmt.Sprintf("/v1/installs/%s/configs/icfg_nonexistent_00000000", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
