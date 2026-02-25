package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func (s *InstallsServiceTestSuite) TestGetComponentOutputsNoDeploy() {
	install := s.createTestInstall()
	ccc := s.testAppConfig.ComponentConfigConnections[0]

	path := fmt.Sprintf("/v1/installs/%s/components/%s/outputs", install.ID, ccc.ComponentID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	// No deploy exists — getInstallComponentLatestDeploy returns an empty deploy,
	// then querying runner jobs by that empty ID returns ErrRecordNotFound → 404.
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
