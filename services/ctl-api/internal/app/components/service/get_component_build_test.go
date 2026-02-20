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

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildSuccess() {
	s.Run("returns build by id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)
		seededBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", s.testApp.ID, cmp.ID, seededBuild.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), seededBuild.ID, response.ID)
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildNotFound() {
	s.Run("nonexistent build id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", s.testApp.ID, cmp.ID, "bld_nonexistent00000000000")
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
