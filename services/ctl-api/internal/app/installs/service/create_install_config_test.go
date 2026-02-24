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

func (s *InstallsServiceTestSuite) TestCreateInstallConfigSuccess() {
	install := s.createTestInstall()

	body := CreateInstallConfigRequest{
		CreateInstallConfigParams: helpers.CreateInstallConfigParams{
			ApprovalOption: app.InstallApprovalOptionApproveAll,
		},
	}

	path := fmt.Sprintf("/v1/installs/%s/configs", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var cfg app.InstallConfig
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &cfg))
	assert.NotEmpty(s.T(), cfg.ID)
	assert.Equal(s.T(), app.InstallApprovalOptionApproveAll, cfg.ApprovalOption)
}
