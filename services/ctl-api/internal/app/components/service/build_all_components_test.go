package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/build"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestBuildAllComponentsSuccess() {
	s.Run("builds all 6 seeded components", func() {
		path := fmt.Sprintf("/v1/apps/%s/components/build-all", s.testApp.ID)
		rr := s.makeRequest(http.MethodPost, path, nil)

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response []*app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		require.Len(s.T(), response, 6, "expected 6 builds for 6 seeded components")

		for _, bld := range response {
			assert.NotEmpty(s.T(), bld.ID)
			assert.Equal(s.T(), app.ComponentBuildStatus("queued"), bld.Status)
		}
	})
}

// ---------------------------------------------------------------------------
// Signals
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestBuildAllComponentsSignals() {
	s.Run("sends OperationBuild signal for each component", func() {

		path := fmt.Sprintf("/v1/apps/%s/components/build-all", s.testApp.ID)
		rr := s.makeRequest(http.MethodPost, path, nil)
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response []*app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		capturedSignals := tests.GetQueueSignals(s.T(), s.deps.DB)
		require.Len(s.T(), capturedSignals, 6, "expected 6 signals")

		// Each signal should be a build signal with a unique BuildID
		buildIDs := map[string]bool{}
		for _, qs := range capturedSignals {
			assert.Equal(s.T(), buildsignal.SignalType, qs.Type, "signal should be component-build")
			sig, ok := qs.Signal.Signal.(*buildsignal.Signal)
			require.True(s.T(), ok, "signal should be *buildsignal.Signal")
			assert.NotEmpty(s.T(), sig.BuildID)
			buildIDs[sig.BuildID] = true
		}
		assert.Len(s.T(), buildIDs, 6, "each signal should have a distinct BuildID")
	})
}

// ---------------------------------------------------------------------------
// Empty app
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestBuildAllComponentsEmptyApp() {
	s.Run("returns empty array for app with no components", func() {
		emptyApp := s.deps.Seeder.CreateApp(s.ctx, s.T())

		path := fmt.Sprintf("/v1/apps/%s/components/build-all", emptyApp.ID)
		rr := s.makeRequest(http.MethodPost, path, nil)

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		// Response may be null or empty array
		body := rr.Body.String()
		assert.True(s.T(), body == "null" || body == "[]",
			"expected null or empty array but got: %s", body)
	})
}
