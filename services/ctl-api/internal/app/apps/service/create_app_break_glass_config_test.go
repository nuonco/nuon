package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestCreateAppBreakGlassConfig tests the CreateAppBreakGlassConfig endpoint.
func (s *AppConfigTypesTestSuite) TestCreateAppBreakGlassConfig() {
	testCases := []struct {
		name         string
		setupFunc    func() CreateAppBreakGlassConfigRequest
		expectedCode int
	}{
		{
			name: "successfully creates break glass config",
			setupFunc: func() CreateAppBreakGlassConfigRequest {
				return CreateAppBreakGlassConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Roles: []AppAWSIAMRoleConfig{
						{
							Name:        "emergency-access",
							DisplayName: "Emergency Access",
							Description: "Break glass emergency access role",
							Policies: []AppAWSIAMPolicyConfig{
								{
									Name:     "emergency-policy",
									Contents: `{"Version": "2012-10-17", "Statement": []}`,
								},
							},
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "fails without app_config_id",
			setupFunc: func() CreateAppBreakGlassConfigRequest {
				return CreateAppBreakGlassConfigRequest{
					Roles: []AppAWSIAMRoleConfig{
						{
							Name:        "emergency-access",
							DisplayName: "Emergency Access",
							Description: "Break glass emergency access role",
							Policies: []AppAWSIAMPolicyConfig{
								{
									Name:     "emergency-policy",
									Contents: `{"Version": "2012-10-17"}`,
								},
							},
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "creates break glass config with minimal role fields",
			setupFunc: func() CreateAppBreakGlassConfigRequest {
				return CreateAppBreakGlassConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					Roles: []AppAWSIAMRoleConfig{
						{
							Name: "minimal-role",
							Policies: []AppAWSIAMPolicyConfig{
								{
									Name:     "policy",
									Contents: `{"Version": "2012-10-17"}`,
								},
							},
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
			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+s.testApp.ID+"/break-glass-configs", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var response app.AppBreakGlassConfig
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), response.ID)
				assert.Equal(s.T(), s.testApp.ID, response.AppID)
				assert.Equal(s.T(), s.testAppConfig.ID, response.AppConfigID)

				var dbConfig app.AppBreakGlassConfig
				err = s.service.DB.Preload("Roles").Preload("Roles.Policies").First(&dbConfig, "id = ?", response.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testApp.ID, dbConfig.AppID)
				assert.NotEmpty(s.T(), dbConfig.Roles)
			}
		})
	}
}
