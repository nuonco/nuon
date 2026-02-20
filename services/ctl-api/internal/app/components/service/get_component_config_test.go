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

func (s *ComponentsServiceTestSuite) TestGetAppComponentConfigSuccess() {
	s.Run("returns config by id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/%s", s.testApp.ID, cmp.ID, ccc.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ComponentConfigConnection
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), ccc.ID, response.ID)
		assert.NotNil(s.T(), response.HelmComponentConfig, "HelmComponentConfig should be populated")
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentConfigNotFound() {
	s.Run("nonexistent config id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/%s", s.testApp.ID, cmp.ID, "ccc_nonexistent00000000000")
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}

func (s *ComponentsServiceTestSuite) TestGetAppComponentConfigWrongComponent() {
	s.Run("config id from different component returns 404", func() {
		helmCmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		tfCmp := s.getSeededComponent(app.ComponentTypeTerraformModule)
		tfCCC := s.getSeededConfigConnection(tfCmp.ID)

		// Use the Terraform config ID but with the Helm component ID in the URL
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/%s", s.testApp.ID, helmCmp.ID, tfCCC.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
