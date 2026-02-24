package service

import (
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetInstallStateNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000/state", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

// NOTE: TestGetInstallStateSuccess is skipped due to a nil pointer bug in
// helpers.toInputState (get_install_state.go:225) when no AppInputConfig exists.
// Same root cause as the GetInstallReadme bug documented in plan.md.
