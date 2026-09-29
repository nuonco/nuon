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

func (s *ComponentsServiceTestSuite) TestDeleteAppComponentSuccess() {
	s.Run("deletes component and sends signal", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)

		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodDelete, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response bool
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.True(s.T(), response)

		var dbComp app.Component
		err = s.deps.DB.WithContext(s.ctx).First(&dbComp, "id = ?", comp.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), app.ComponentStatusDeleteQueued, dbComp.Status)
		assert.Equal(s.T(), "delete has been queued and waiting", dbComp.StatusDescription)

		capturedSignals := tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, comp.ID)
		require.Len(s.T(), capturedSignals, 1, "expected 1 signal")

		assert.Equal(s.T(), comp.ID, capturedSignals[0].OwnerID, "signal should target the deleted component")
		assert.Equal(s.T(), componentdelete.SignalType, capturedSignals[0].Type)
	})
}

func (s *ComponentsServiceTestSuite) TestDeleteAppComponentRejectsActiveConfigComponent() {
	s.Run("rejects deletion of component in active app config", func() {
		seededComponentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID

		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, seededComponentID)
		rr := s.makeRequest(http.MethodDelete, path, nil)

		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})
}

func (s *ComponentsServiceTestSuite) TestDeleteAppComponentNotFound() {
	s.Run("nonexistent component id", func() {
		path := fmt.Sprintf("/v1/apps/%s/components/%s", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodDelete, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)

		capturedSignals := tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, "cmp_nonexistent00000000000")
		assert.Len(s.T(), capturedSignals, 0, "should not send signal when component not found")
	})
}
