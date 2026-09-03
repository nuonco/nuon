package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestUpdateInstallName() {
	install := s.createTestInstall()

	body := map[string]interface{}{
		"name": "updated-name",
	}

	path := fmt.Sprintf("/v1/installs/%s", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), "updated-name", resp.Name)

	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	require.Len(s.T(), captured, 2)
	assert.ElementsMatch(s.T(), []string{"updated", "appconfig-updated"}, []string{
		string(captured[0].Type),
		string(captured[1].Type),
	})
}

func (s *InstallsServiceTestSuite) TestUpdateInstallNotFound() {
	body := map[string]interface{}{
		"name": "updated-name",
	}

	rr := s.makeRequest(http.MethodPatch, "/v1/installs/ins_nonexistent_00000000", body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
