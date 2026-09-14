package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
)

func (s *InstallsServiceTestSuite) TestGetInstallUpdatesPaginatesMergedFeed() {
	install := s.createTestInstall()
	firstValue := "first"
	secondValue := "second"
	s.deps.Seeder.CreateInstallInputs(
		s.ctx,
		s.T(),
		install.ID,
		s.testAppConfig.InputConfig.ID,
		map[string]*string{"license": &firstValue},
	)
	s.deps.Seeder.CreateInstallInputs(
		s.ctx,
		s.T(),
		install.ID,
		s.testAppConfig.InputConfig.ID,
		map[string]*string{"license": &secondValue},
	)

	firstPage := s.makeRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/installs/%s/updates?limit=1&page=0", install.ID),
		nil,
	)
	require.Equal(s.T(), http.StatusOK, firstPage.Code, firstPage.Body.String())
	var firstResponse InstallUpdatesResponse
	require.NoError(s.T(), json.Unmarshal(firstPage.Body.Bytes(), &firstResponse))
	require.Len(s.T(), firstResponse.Updates, 1)
	require.True(s.T(), firstResponse.HasMore)

	secondPage := s.makeRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/installs/%s/updates?limit=1&page=1", install.ID),
		nil,
	)
	require.Equal(s.T(), http.StatusOK, secondPage.Code, secondPage.Body.String())
	var secondResponse InstallUpdatesResponse
	require.NoError(s.T(), json.Unmarshal(secondPage.Body.Bytes(), &secondResponse))
	require.Len(s.T(), secondResponse.Updates, 1)
	require.NotEqual(s.T(), firstResponse.Updates[0].ID, secondResponse.Updates[0].ID)
}
