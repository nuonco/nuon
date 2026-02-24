package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentLatestConfigSuccess() {
	s.Run("returns latest config for seeded component", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/latest", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ComponentConfigConnection
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), ccc.ID, response.ID)
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentLatestConfigNotFound() {
	s.Run("no configs for fresh component", func() {
		// Create a fresh component with no config connections
		freshComp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/latest", s.testApp.ID, freshComp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}

func (s *ComponentsServiceTestSuite) TestGetAppComponentLatestConfigWrongApp() {
	s.Run("correct component but wrong app id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		// Use a different app ID in the URL
		otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/latest", otherApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
