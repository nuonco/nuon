package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestUpdateApp tests the UpdateApp endpoint.
func (s *AppCRUDTestSuite) TestUpdateApp() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		updateReq    UpdateAppRequest
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "update app name returns 200",
			setupFunc: func() string {
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "original-name",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err := s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				return testApp.ID
			},
			updateReq: UpdateAppRequest{
				Name: "updated-name",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(appID string) {
				var dbApp app.App
				err := s.service.DB.First(&dbApp, "id = ?", appID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "updated-name", dbApp.Name)
			},
		},
		{
			name: "update app description returns 200",
			setupFunc: func() string {
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-desc",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err := s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				return testApp.ID
			},
			updateReq: UpdateAppRequest{
				Name:        "test-app-desc",
				Description: "New description",
			},
			expectedCode: http.StatusOK,
			validateFunc: func(appID string) {
				var dbApp app.App
				err := s.service.DB.First(&dbApp, "id = ?", appID).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), dbApp.Description.Valid)
				assert.Equal(s.T(), "New description", dbApp.Description.String)
			},
		},
		{
			name: "update non-existent app returns 404",
			setupFunc: func() string {
				return domains.NewAppID()
			},
			updateReq: UpdateAppRequest{
				Name: "updated-name",
			},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			appID := tc.setupFunc()
			rr := s.makeRequest(http.MethodPatch, "/v1/apps/"+appID, tc.updateReq)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				var response models.AppApp
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				tc.validateFunc(appID)
			}
		})
	}
}
