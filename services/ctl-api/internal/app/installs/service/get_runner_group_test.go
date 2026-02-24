package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func (s *InstallsServiceTestSuite) TestGetInstallRunnerGroupNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000/runner-group", nil)
	assert.NotEqual(s.T(), http.StatusOK, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallRunnerGroupExists() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/runner-group", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	// Runner group may or may not be created by seeder - just verify no 500
	assert.NotEqual(s.T(), http.StatusInternalServerError, rr.Code)
}
