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

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildsSuccess() {
	s.Run("returns seeded build", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)
		seededBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response []app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		require.GreaterOrEqual(s.T(), len(response), 1, "expected at least 1 build")

		var found bool
		for _, bld := range response {
			if bld.ID == seededBuild.ID {
				found = true
				break
			}
		}
		assert.True(s.T(), found, "seeded build should be in the response")
	})
}

// ---------------------------------------------------------------------------
// Empty cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildsEmpty() {
	s.Run("returns empty array for component with no builds", func() {
		// Create a fresh component with a config connection but no builds
		freshComp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)
		s.deps.Seeder.CreateDockerBuildComponentConfigConnection(s.ctx, s.T(), freshComp.ID, s.testAppConfig.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds", s.testApp.ID, freshComp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response []app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Len(s.T(), response, 0, "expected no builds")
	})
}
