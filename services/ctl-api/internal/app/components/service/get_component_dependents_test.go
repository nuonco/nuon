package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentDependentsSuccess() {
	s.Run("returns empty children for component with no dependents", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/dependents", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response ComponentChildren
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Len(s.T(), response.Children, 0)
	})

	s.Run("returns dependent when B depends on A", func() {
		// A is the dependency, B is the dependent.
		// The graph edge goes A → B, so BFS from A finds B.
		cmpA := s.getSeededComponent(app.ComponentTypeHelmChart)
		cmpB := s.getSeededComponent(app.ComponentTypeTerraformModule)

		// Set B's config connection to declare A as a dependency
		cccB := s.getSeededConfigConnection(cmpB.ID)
		res := s.deps.DB.Model(&app.ComponentConfigConnection{}).
			Where("id = ?", cccB.ID).
			Update("component_dependency_ids", pq.StringArray{cmpA.ID})
		require.NoError(s.T(), res.Error)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/dependents", s.testApp.ID, cmpA.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response ComponentChildren
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		require.Len(s.T(), response.Children, 1, "expected B as dependent of A")
		assert.Equal(s.T(), cmpB.ID, response.Children[0].ID)
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentDependentsNotFound() {
	s.Run("nonexistent component id", func() {
		path := fmt.Sprintf("/v1/apps/%s/components/%s/dependents", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}

func (s *ComponentsServiceTestSuite) TestGetAppComponentDependentsNotInActiveConfig() {
	s.Run("component not in active app config ComponentIDs", func() {
		// Create a fresh component that is NOT in the active app config's ComponentIDs
		freshComp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/dependents", s.testApp.ID, freshComp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		// Should fail because freshComp is not in appCfg.ComponentIDs
		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
