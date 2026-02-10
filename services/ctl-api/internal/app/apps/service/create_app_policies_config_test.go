package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestCreateAppPoliciesConfig tests the CreateAppPoliciesConfig endpoint.
func (s *AppConfigTypesTestSuite) TestCreateAppPoliciesConfig() {
	testCases := []struct {
		name         string
		setupFunc    func() CreateAppPoliciesConfigRequest
		expectedCode int
	}{
		{
			name: "successfully creates policies config",
			setupFunc: func() CreateAppPoliciesConfigRequest {
				return CreateAppPoliciesConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Policies: []AppPolicyConfig{
						{
							Type:        config.AppPolicyTypeKubernetesCluster,
							Engine:      config.AppPolicyEngineOPA,
							Name:        "test-policy",
							Description: "Test policy description",
							Contents:    "package test\n\ndefault allow = false",
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "fails without app_config_id",
			setupFunc: func() CreateAppPoliciesConfigRequest {
				return CreateAppPoliciesConfigRequest{
					Policies: []AppPolicyConfig{
						{
							Type:     config.AppPolicyTypeKubernetesCluster,
							Contents: "package test",
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "creates policies config with minimal fields",
			setupFunc: func() CreateAppPoliciesConfigRequest {
				return CreateAppPoliciesConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Policies: []AppPolicyConfig{
						{
							Type: config.AppPolicyTypeKubernetesCluster,
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.setupFunc()
			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+s.testApp.ID+"/policies-configs", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var response app.AppPoliciesConfig
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), response.ID)
				assert.Equal(s.T(), s.testApp.ID, response.AppID)
				assert.Equal(s.T(), s.testAppConfig.ID, response.AppConfigID)

				var dbConfig app.AppPoliciesConfig
				err = s.service.DB.Preload("Policies").First(&dbConfig, "id = ?", response.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testApp.ID, dbConfig.AppID)
			}
		})
	}
}
