package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetInstallInputsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.InstallInputs
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}

func (s *InstallsServiceTestSuite) TestGetInstallInputsCurrentNoInputs() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/inputs/current", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
