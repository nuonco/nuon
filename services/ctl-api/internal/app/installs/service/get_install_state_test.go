package service

import (
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetInstallStateNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000/state", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
