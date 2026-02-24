package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallStackByStackIDSuccess() {
	install := s.createTestInstall()

	var stack app.InstallStack
	res := s.deps.DB.Where("install_id = ?", install.ID).First(&stack)
	require.NoError(s.T(), res.Error)

	path := fmt.Sprintf("/v1/installs/stacks/%s", stack.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.InstallStack
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), stack.ID, resp.ID)
}

func (s *InstallsServiceTestSuite) TestGetInstallStackByStackIDNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/stacks/stk_nonexistent_00000000", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
