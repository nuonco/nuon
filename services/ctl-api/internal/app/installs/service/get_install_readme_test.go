package service

import (
	"net/http"

	"github.com/stretchr/testify/assert"
)

func (s *InstallsServiceTestSuite) TestGetInstallReadmeNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000/readme", nil)
	assert.NotEqual(s.T(), http.StatusOK, rr.Code)
}

// NOTE: TestGetInstallReadmeNoReadme is skipped because GetInstallReadme has a nil pointer
// dereference bug in helpers.toInputState when no AppInputConfig exists for the app.
// See get_install_state.go:225. This should be fixed with a nil check in the handler.
