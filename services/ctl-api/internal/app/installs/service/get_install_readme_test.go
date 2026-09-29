package service

import (
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetInstallReadmeNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000/readme", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
