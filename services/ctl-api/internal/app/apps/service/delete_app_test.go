package service

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// TestDeleteApp tests the DeleteApp endpoint.
func (s *AppCRUDTestSuite) TestDeleteApp() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "delete app with no installs sets status to DeleteQueued and returns 200",
			setupFunc: func() string {
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-delete",
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
			expectedCode: http.StatusOK,
			validateFunc: func(appID string) {
				var dbApp app.App
				err := s.service.DB.First(&dbApp, "id = ?", appID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.AppStatusDeleteQueued, dbApp.Status)
				assert.Equal(s.T(), "delete has been queued and waiting", dbApp.StatusDescription)
			},
		},
		{
			name: "delete non-existent app returns 404",
			setupFunc: func() string {
				return domains.NewAppID()
			},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
		{
			name: "delete app with active installs returns 500",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-with-installs",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err := s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				install := &app.Install{
					ID:    domains.NewInstallID(),
					Name:  "test-install",
					AppID: testApp.ID,
				}
				err = s.service.DB.WithContext(ctx).
					Omit("app_config_id", "app_sandbox_config_id", "app_runner_config_id").
					Create(install).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", install.ID)
				})

				sandboxRun := &app.InstallSandboxRun{
					ID:        domains.NewSandboxRunID(),
					InstallID: install.ID,
					Status:    app.SandboxRunStatusProvisioning,
				}
				err = s.service.DB.WithContext(ctx).
					Omit("app_sandbox_config_id", "install_sandbox_id", "install_workflow_id").
					Create(sandboxRun).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.InstallSandboxRun{}, "id = ?", sandboxRun.ID)
				})

				return testApp.ID
			},
			expectedCode: http.StatusInternalServerError,
			validateFunc: func(appID string) {
				var dbApp app.App
				err := s.service.DB.First(&dbApp, "id = ?", appID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.AppStatusProvisioning, dbApp.Status)
			},
		},
		{
			name: "delete app with deprovisioned installs returns 200",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-deprovisioned",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusProvisioning,
				}
				err := s.service.DB.Create(testApp).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				install := &app.Install{
					ID:    domains.NewInstallID(),
					Name:  "test-install-deprovisioned",
					AppID: testApp.ID,
				}
				err = s.service.DB.WithContext(ctx).
					Omit("app_config_id", "app_sandbox_config_id", "app_runner_config_id").
					Create(install).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", install.ID)
				})

				sandboxRun := &app.InstallSandboxRun{
					ID:        domains.NewSandboxRunID(),
					InstallID: install.ID,
					Status:    app.SandboxRunStatusDeprovisioned,
				}
				err = s.service.DB.WithContext(ctx).
					Omit("app_sandbox_config_id", "install_sandbox_id", "install_workflow_id").
					Create(sandboxRun).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.InstallSandboxRun{}, "id = ?", sandboxRun.ID)
				})

				return testApp.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(appID string) {
				var dbApp app.App
				err := s.service.DB.First(&dbApp, "id = ?", appID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.AppStatusDeleteQueued, dbApp.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			appID := tc.setupFunc()
			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+appID, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(appID)
			}
		})
	}
}
