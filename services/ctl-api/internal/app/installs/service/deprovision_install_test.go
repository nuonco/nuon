package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestDeprovisionInstallSuccess() {
	install := s.createTestInstallWithActiveRunner()

	path := fmt.Sprintf("/v1/installs/%s/deprovision", install.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	require.Len(s.T(), captured, 1)
	_ = captured[0] // signal type check via .Type

	assert.Equal(s.T(), "ExecuteFlow-type", string(captured[0].Type))
}

func (s *InstallsServiceTestSuite) TestDeprovisionInstallPlanOnly() {
	install := s.createTestInstallWithActiveRunner()

	body := DeprovisionInstallRequest{PlanOnly: true}

	path := fmt.Sprintf("/v1/installs/%s/deprovision", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)
}

func (s *InstallsServiceTestSuite) TestDeprovisionInstallWithoutActiveRunner() {
	install := s.deps.Seeder.CreateInstall(s.ctx, s.T(), s.testApp)
	path := fmt.Sprintf("/v1/installs/%s/deprovision", install.ID)

	rr := s.makeRequest(http.MethodPost, path, nil)

	require.Equal(s.T(), http.StatusConflict, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "runner is not currently active")
	var count int64
	require.NoError(s.T(), s.deps.DB.Model(&app.Workflow{}).
		Where(app.Workflow{OwnerID: install.ID}).
		Count(&count).Error)
	require.Zero(s.T(), count)
}

func (s *InstallsServiceTestSuite) TestDeprovisionInstallNotFound() {
	rr := s.makeRequest(http.MethodPost, "/v1/installs/ins_nonexistent_00000000/deprovision", nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
