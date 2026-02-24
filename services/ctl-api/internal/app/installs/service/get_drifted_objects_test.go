package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetDriftedObjectsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/drifted-objects", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var objs []app.DriftedObject
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &objs))
	assert.Empty(s.T(), objs)
}
