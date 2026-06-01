package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	configcreated "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/configcreated"
	updatecomptype "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/updatecomponenttype"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppJobConfigSuccess() {
	s.Run("creates config with image_url and tag", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeJob)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/job", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateJobComponentConfigRequest{
			ImageURL:    "ubuntu",
			Tag:         "latest",
			AppConfigID: s.testAppConfig.ID,
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.JobComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "ubuntu", response.ImageURL)
		assert.Equal(s.T(), "latest", response.Tag)
	})
}

func (s *ComponentsServiceTestSuite) TestCreateAppJobConfigWithOptionalFields() {
	s.Run("creates config with cmd, args, and env_vars", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeJob)

		envVal := "world"
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/job", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateJobComponentConfigRequest{
			ImageURL:    "ubuntu",
			Tag:         "22.04",
			AppConfigID: s.testAppConfig.ID,
			Cmd:         []string{"echo", "hello"},
			Args:        []string{"--verbose"},
			EnvVars:     map[string]*string{"HELLO": &envVal},
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.JobComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "22.04", response.Tag)
		assert.Equal(s.T(), []string{"echo", "hello"}, []string(response.Cmd))
		assert.Equal(s.T(), []string{"--verbose"}, []string(response.Args))
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppJobConfigValidationErrors() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeJob)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/job", s.testApp.ID, comp.ID)

	testCases := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing image_url",
			body: map[string]interface{}{
				"tag":           "latest",
				"app_config_id": s.testAppConfig.ID,
			},
		},
		{
			name: "missing tag",
			body: map[string]interface{}{
				"image_url":     "ubuntu",
				"app_config_id": s.testAppConfig.ID,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodPost, path, tc.body)

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// Signals
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppJobConfigSignals() {
	s.Run("sends OperationConfigCreated and OperationUpdateComponentType signals", func() {

		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeJob)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/job", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateJobComponentConfigRequest{
			ImageURL:    "ubuntu",
			Tag:         "latest",
			AppConfigID: s.testAppConfig.ID,
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		capturedSignals := tests.GetQueueSignals(s.T(), s.deps.DB)
		require.Len(s.T(), capturedSignals, 2, "expected 2 signals")

		assert.Equal(s.T(), configcreated.SignalType, capturedSignals[0].Type)

		assert.Equal(s.T(), updatecomptype.SignalType, capturedSignals[1].Type)
		sig1, ok := capturedSignals[1].Signal.Signal.(*updatecomptype.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), app.ComponentTypeJob, sig1.ComponentType)
	})
}
