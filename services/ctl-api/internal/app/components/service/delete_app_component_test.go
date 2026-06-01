package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componentdelete "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/delete"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestDeleteAppComponentSuccess() {
	s.Run("deletes component and sends signal", func() {
		// Reset mock

		// Create a component that is NOT in the current app config's ComponentIDs.
		// The delete endpoint rejects components that are part of the active config,
		// so deletion only works for components removed from the config.
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)

		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodDelete, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		// Verify response is true
		var response bool
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.True(s.T(), response)

		// Verify component status is set to delete_queued (not actually soft-deleted)
		var dbComp app.Component
		err = s.deps.DB.WithContext(s.ctx).First(&dbComp, "id = ?", comp.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), app.ComponentStatusDeleteQueued, dbComp.Status)
		assert.Equal(s.T(), "delete has been queued and waiting", dbComp.StatusDescription)

		// Verify OperationDelete signal was sent
		capturedSignals := tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, comp.ID)
		require.Len(s.T(), capturedSignals, 1, "expected 1 signal")

		assert.Equal(s.T(), comp.ID, capturedSignals[0].OwnerID, "signal should target the deleted component")
		assert.Equal(s.T(), componentdelete.SignalType, capturedSignals[0].Type)
	})
}

func (s *ComponentsServiceTestSuite) TestDeleteAppComponentRejectsActiveConfigComponent() {
	s.Run("rejects deletion of component in active app config", func() {
		// Use a pre-seeded component from the full app config — should be rejected
		seededComponentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID

		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, seededComponentID)
		rr := s.makeRequest(http.MethodDelete, path, nil)

		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestDeleteAppComponentNotFound() {
	s.Run("nonexistent component id", func() {
		// Reset mock

		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodDelete, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)

		// Verify no signal was sent
		capturedSignals := tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, "cmp_nonexistent00000000000")
		assert.Len(s.T(), capturedSignals, 0, "should not send signal when component not found")
	})
}
