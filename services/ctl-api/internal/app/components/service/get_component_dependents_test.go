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
		cmpA := s.getSeededComponent(app.ComponentTypeHelmChart)
		cmpB := s.getSeededComponent(app.ComponentTypeTerraformModule)

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
		freshComp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/dependents", s.testApp.ID, freshComp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
