package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestCreateAppPermissionsConfig tests the CreateAppPermissionsConfig endpoint.
func (s *AppConfigTypesTestSuite) TestCreateAppPermissionsConfig() {
	testCases := []struct {
		name         string
		setupFunc    func() CreateAppPermissionsConfigRequest
		expectedCode int
	}{
		{
			name: "successfully creates permissions config with all required roles",
			setupFunc: func() CreateAppPermissionsConfigRequest {
				return CreateAppPermissionsConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					ProvisionRole: AppAWSIAMRoleConfig{
						Name:        "provision-role",
						DisplayName: "Provision Role",
						Description: "Role for provisioning infrastructure",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "provision-policy",
								Contents: `{"Version": "2012-10-17", "Statement": []}`,
							},
						},
					},
					DeprovisionRole: AppAWSIAMRoleConfig{
						Name:        "deprovision-role",
						DisplayName: "Deprovision Role",
						Description: "Role for deprovisioning infrastructure",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "deprovision-policy",
								Contents: `{"Version": "2012-10-17", "Statement": []}`,
							},
						},
					},
					MaintenanceRole: AppAWSIAMRoleConfig{
						Name:        "maintenance-role",
						DisplayName: "Maintenance Role",
						Description: "Role for maintenance operations",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "maintenance-policy",
								Contents: `{"Version": "2012-10-17", "Statement": []}`,
							},
						},
					},
				}
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "fails without app_config_id",
			setupFunc: func() CreateAppPermissionsConfigRequest {
				return CreateAppPermissionsConfigRequest{
					ProvisionRole: AppAWSIAMRoleConfig{
						Name:        "provision-role",
						DisplayName: "Provision Role",
						Description: "Role for provisioning",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "policy",
								Contents: `{"Version": "2012-10-17"}`,
							},
						},
					},
					DeprovisionRole: AppAWSIAMRoleConfig{
						Name:        "deprovision-role",
						DisplayName: "Deprovision Role",
						Description: "Role for deprovisioning",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "policy",
								Contents: `{"Version": "2012-10-17"}`,
							},
						},
					},
					MaintenanceRole: AppAWSIAMRoleConfig{
						Name:        "maintenance-role",
						DisplayName: "Maintenance Role",
						Description: "Role for maintenance",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "policy",
								Contents: `{"Version": "2012-10-17"}`,
							},
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "fails without provision role",
			setupFunc: func() CreateAppPermissionsConfigRequest {
				return CreateAppPermissionsConfigRequest{
					AppConfigID: s.testAppConfig.ID,
					DeprovisionRole: AppAWSIAMRoleConfig{
						Name:        "deprovision-role",
						DisplayName: "Deprovision Role",
						Description: "Role for deprovisioning",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "policy",
								Contents: `{"Version": "2012-10-17"}`,
							},
						},
					},
					MaintenanceRole: AppAWSIAMRoleConfig{
						Name:        "maintenance-role",
						DisplayName: "Maintenance Role",
						Description: "Role for maintenance",
						Policies: []AppAWSIAMPolicyConfig{
							{
								Name:     "policy",
								Contents: `{"Version": "2012-10-17"}`,
							},
						},
					},
				}
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.setupFunc()
			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+s.testApp.ID+"/permissions-configs", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var response app.AppPermissionsConfig
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)

				assert.NotEmpty(s.T(), response.ID)
				assert.Equal(s.T(), s.testApp.ID, response.AppID)
				assert.Equal(s.T(), s.testAppConfig.ID, response.AppConfigID)

				var dbConfig app.AppPermissionsConfig
				err = s.service.DB.Preload("Roles").Preload("Roles.Policies").First(&dbConfig, "id = ?", response.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), s.testApp.ID, dbConfig.AppID)
				assert.GreaterOrEqual(s.T(), len(dbConfig.Roles), 3)
			}
		})
	}
}
