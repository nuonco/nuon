package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetAvailableRolesEmptyOutputs() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/available-roles?principal_type=component&operation_type=deploy", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp AvailableRolesResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Empty(s.T(), resp.Roles)
}

func (s *InstallsServiceTestSuite) TestGetAvailableRolesMissingPrincipalType() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/available-roles?operation_type=deploy", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetAvailableRolesMissingOperationType() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/available-roles?principal_type=component", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetAvailableRolesInvalidPrincipalType() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/available-roles?principal_type=invalid&operation_type=deploy", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}
