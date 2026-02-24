package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppExternalImageConfigSuccess() {
	s.Run("creates config with image_url and tag", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeExternalImage)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/external-image", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateExternalImageComponentConfigRequest{
			ImageURL:    "nginx",
			Tag:         "latest",
			AppConfigID: s.testAppConfig.ID,
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.ExternalImageComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "nginx", response.ImageURL)
		assert.Equal(s.T(), "latest", response.Tag)
	})
}

func (s *ComponentsServiceTestSuite) TestCreateAppExternalImageConfigWithAWSECR() {
	s.Run("creates config with AWS ECR image config", func() {
		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeExternalImage)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/external-image", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateExternalImageComponentConfigRequest{
			ImageURL:    "123456789.dkr.ecr.us-west-2.amazonaws.com/my-repo",
			Tag:         "v1.0.0",
			AppConfigID: s.testAppConfig.ID,
			AWSECRImageConfig: &awsECRImageConfigRequest{
				IAMRoleARN: "arn:aws:iam::123456789:role/ecr-access",
				AWSRegion:  "us-west-2",
			},
		})

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var response app.ExternalImageComponentConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "v1.0.0", response.Tag)
		assert.NotNil(s.T(), response.AWSECRImageConfig, "AWSECRImageConfig should be populated")
	})
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppExternalImageConfigValidationErrors() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeExternalImage)
	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/external-image", s.testApp.ID, comp.ID)

	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
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
				"image_url":     "nginx",
				"app_config_id": s.testAppConfig.ID,
			},
		},
		{
			name:    "invalid JSON",
			rawBody: "{invalid json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var rr *httptest.ResponseRecorder
			if tc.rawBody != "" {
				rr = s.makeRawRequest(http.MethodPost, path, tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodPost, path, tc.body)
			}

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

func (s *ComponentsServiceTestSuite) TestCreateAppExternalImageConfigSignals() {
	s.Run("sends OperationConfigCreated and OperationUpdateComponentType signals", func() {
		s.mockEvClient.Reset()

		comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeExternalImage)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/external-image", s.testApp.ID, comp.ID)
		rr := s.makeRequest(http.MethodPost, path, CreateExternalImageComponentConfigRequest{
			ImageURL:    "nginx",
			Tag:         "latest",
			AppConfigID: s.testAppConfig.ID,
		})
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		capturedSignals := s.mockEvClient.GetSignals()
		require.Len(s.T(), capturedSignals, 2, "expected 2 signals")

		sig0, ok := capturedSignals[0].Signal.(*signals.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), signals.OperationConfigCreated, sig0.Type)

		sig1, ok := capturedSignals[1].Signal.(*signals.Signal)
		require.True(s.T(), ok)
		assert.Equal(s.T(), signals.OperationUpdateComponentType, sig1.Type)
		assert.Equal(s.T(), app.ComponentTypeExternalImage, sig1.ComponentType)
	})
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateAppExternalImageConfigNotFound() {
	s.Run("nonexistent component id", func() {
		path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/external-image", s.testApp.ID, "cmp_nonexistent00000000000")
		rr := s.makeRequest(http.MethodPost, path, CreateExternalImageComponentConfigRequest{
			ImageURL:    "nginx",
			Tag:         "latest",
			AppConfigID: s.testAppConfig.ID,
		})

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
