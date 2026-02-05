package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// GetOrgStatsTestService holds all fx-injected dependencies for org stats endpoint tests.
type GetOrgStatsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
}

// GetOrgStatsTestSuite is the testify suite for GetOrgStats endpoint.
type GetOrgStatsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service GetOrgStatsTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestGetOrgStatsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(GetOrgStatsTestSuite))
}

func (s *GetOrgStatsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)

	s.app.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *GetOrgStatsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create test router with standard middlewares and org context
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *GetOrgStatsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetOrgStatsTestSuite) setupTestData() {
	// Create test account
	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	// Create test org with account context (required by BeforeCreate hook)
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:   domains.NewOrgID(),
		Name: "test-org-" + domains.NewOrgID(),
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *GetOrgStatsTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetOrgStatsTestSuite) TestGetOrgStats() {
	testCases := []struct {
		name              string
		setupFunc         func()
		expectedAppCount  int64
		expectedInstCount int64
		expectedInstNames []string
	}{
		{
			name: "returns zeros when no apps or installs",
			setupFunc: func() {
				// No setup needed - tests empty org
			},
			expectedAppCount:  0,
			expectedInstCount: 0,
			expectedInstNames: []string{},
		},
		{
			name: "returns correct app count",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				app1 := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-1",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}
				app2 := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app-2",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}

				err := s.service.DB.WithContext(ctx).Create(app1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app1.ID)
				})

				err = s.service.DB.WithContext(ctx).Create(app2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", app2.ID)
				})
			},
			expectedAppCount:  2,
			expectedInstCount: 0,
			expectedInstNames: []string{},
		},
		{
			name: "returns correct install count",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create app first (installs require an app)
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(testApp).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				// Create installs
				install1 := &app.Install{
					ID:          domains.NewInstallID(),
					Name:        "test-install-1",
					OrgID:       s.testOrg.ID,
					AppID:       testApp.ID,
					CreatedByID: s.testAcc.ID,
				}
				install2 := &app.Install{
					ID:          domains.NewInstallID(),
					Name:        "test-install-2",
					OrgID:       s.testOrg.ID,
					AppID:       testApp.ID,
					CreatedByID: s.testAcc.ID,
				}

				err = s.service.DB.WithContext(ctx).Create(install1).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", install1.ID)
				})

				err = s.service.DB.WithContext(ctx).Create(install2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", install2.ID)
				})
			},
			expectedAppCount:  1,
			expectedInstCount: 2,
			expectedInstNames: []string{"test-install-1", "test-install-2"},
		},
		{
			name: "returns install names array",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create app
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "test-app",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(testApp).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				// Create install with specific name
				install := &app.Install{
					ID:          domains.NewInstallID(),
					Name:        "production-install",
					OrgID:       s.testOrg.ID,
					AppID:       testApp.ID,
					CreatedByID: s.testAcc.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(install).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", install.ID)
				})
			},
			expectedAppCount:  1,
			expectedInstCount: 1,
			expectedInstNames: []string{"production-install"},
		},
		{
			name: "only counts apps and installs for current org",
			setupFunc: func() {
				// Create second account and org
				acc2 := &app.Account{
					ID:          domains.NewAccountID(),
					Email:       "other@example.com",
					Subject:     "other-subject",
					AccountType: app.AccountTypeAuth0,
				}
				err := s.service.DB.Create(acc2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Account{}, "id = ?", acc2.ID)
				})

				ctx2 := context.Background()
				ctx2 = cctx.SetAccountContext(ctx2, acc2)

				org2 := &app.Org{
					ID:   domains.NewOrgID(),
					Name: "other-org-" + domains.NewOrgID(),
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/foo",
					},
				}
				err = s.service.DB.WithContext(ctx2).Create(org2).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org2.ID)
				})

				// Create app and install for test org
				ctx1 := context.Background()
				ctx1 = cctx.SetAccountContext(ctx1, s.testAcc)

				myApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "my-app",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}
				err = s.service.DB.WithContext(ctx1).Create(myApp).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", myApp.ID)
				})

				myInstall := &app.Install{
					ID:          domains.NewInstallID(),
					Name:        "my-install",
					OrgID:       s.testOrg.ID,
					AppID:       myApp.ID,
					CreatedByID: s.testAcc.ID,
				}
				err = s.service.DB.WithContext(ctx1).Create(myInstall).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", myInstall.ID)
				})

				// Create app and install for other org (should NOT be counted)
				otherApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "other-app",
					OrgID:       org2.ID,
					CreatedByID: acc2.ID,
					Status:      app.AppStatusActive,
				}
				err = s.service.DB.WithContext(ctx2).Create(otherApp).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", otherApp.ID)
				})

				otherInstall := &app.Install{
					ID:          domains.NewInstallID(),
					Name:        "other-install",
					OrgID:       org2.ID,
					AppID:       otherApp.ID,
					CreatedByID: acc2.ID,
				}
				err = s.service.DB.WithContext(ctx2).Create(otherInstall).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", otherInstall.ID)
				})
			},
			expectedAppCount:  1,
			expectedInstCount: 1,
			expectedInstNames: []string{"my-install"},
		},
		{
			name: "multiple installs with different names",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create app
				testApp := &app.App{
					ID:          domains.NewAppID(),
					Name:        "multi-install-app",
					OrgID:       s.testOrg.ID,
					CreatedByID: s.testAcc.ID,
					Status:      app.AppStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(testApp).Error
				require.NoError(s.T(), err)
				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
				})

				// Create multiple installs with different names
				installNames := []string{
					"production-install",
					"staging-install",
					"development-install",
				}

				for _, name := range installNames {
					install := &app.Install{
						ID:          domains.NewInstallID(),
						Name:        name,
						OrgID:       s.testOrg.ID,
						AppID:       testApp.ID,
						CreatedByID: s.testAcc.ID,
					}
					err := s.service.DB.WithContext(ctx).Create(install).Error
					require.NoError(s.T(), err)
					installID := install.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", installID)
					})
				}
			},
			expectedAppCount:  1,
			expectedInstCount: 3,
			expectedInstNames: []string{
				"production-install",
				"staging-install",
				"development-install",
			},
		},
		{
			name: "multiple apps with multiple installs",
			setupFunc: func() {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create three apps
				for i := 1; i <= 3; i++ {
					testApp := &app.App{
						ID:          domains.NewAppID(),
						Name:        fmt.Sprintf("app-%d", i),
						OrgID:       s.testOrg.ID,
						CreatedByID: s.testAcc.ID,
						Status:      app.AppStatusActive,
					}
					err := s.service.DB.WithContext(ctx).Create(testApp).Error
					require.NoError(s.T(), err)
					appID := testApp.ID
					s.T().Cleanup(func() {
						s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", appID)
					})

					// Create two installs per app
					for j := 1; j <= 2; j++ {
						install := &app.Install{
							ID:          domains.NewInstallID(),
							Name:        fmt.Sprintf("app-%d-install-%d", i, j),
							OrgID:       s.testOrg.ID,
							AppID:       testApp.ID,
							CreatedByID: s.testAcc.ID,
						}
						err := s.service.DB.WithContext(ctx).Create(install).Error
						require.NoError(s.T(), err)
						installID := install.ID
						s.T().Cleanup(func() {
							s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", installID)
						})
					}
				}
			},
			expectedAppCount:  3,
			expectedInstCount: 6,
			expectedInstNames: []string{
				"app-1-install-1",
				"app-1-install-2",
				"app-2-install-1",
				"app-2-install-2",
				"app-3-install-1",
				"app-3-install-2",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			tc.setupFunc()

			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/stats")

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			// Parse response
			var response OrgStatsResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			// Verify counts
			assert.Equal(s.T(), tc.expectedAppCount, response.AppCount,
				"app_count mismatch")
			assert.Equal(s.T(), tc.expectedInstCount, response.InstallCount,
				"install_count mismatch")

			// Verify install names array
			require.NotNil(s.T(), response.InstallNames, "install_names should not be nil")
			assert.Equal(s.T(), len(tc.expectedInstNames), len(response.InstallNames),
				"install_names length mismatch")

			// Verify all expected install names are present
			if len(tc.expectedInstNames) > 0 {
				for _, expectedName := range tc.expectedInstNames {
					assert.Contains(s.T(), response.InstallNames, expectedName,
						"expected install name not found: %s", expectedName)
				}
			}
		})
	}
}

func (s *GetOrgStatsTestSuite) TestGetOrgStatsVerifiesDatabaseState() {
	// Setup test data
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	testApp := &app.App{
		ID:          domains.NewAppID(),
		Name:        "verify-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusActive,
	}
	err := s.service.DB.WithContext(ctx).Create(testApp).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)

	install := &app.Install{
		ID:          domains.NewInstallID(),
		Name:        "verify-install",
		OrgID:       s.testOrg.ID,
		AppID:       testApp.ID,
		CreatedByID: s.testAcc.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(install).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.Install{}, "id = ?", install.ID)

	// Make API request
	rr := s.makeRequest(http.MethodGet, "/v1/orgs/current/stats")
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response OrgStatsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Verify database state matches response
	var dbAppCount int64
	err = s.service.DB.Model(&app.App{}).Where("org_id = ?", s.testOrg.ID).Count(&dbAppCount).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), dbAppCount, response.AppCount,
		"response app_count should match database count")

	var dbInstallCount int64
	err = s.service.DB.Model(&app.Install{}).Where("org_id = ?", s.testOrg.ID).Count(&dbInstallCount).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), dbInstallCount, response.InstallCount,
		"response install_count should match database count")

	// Verify install names in database
	var dbInstallNames []string
	err = s.service.DB.Model(&app.Install{}).
		Where("org_id = ?", s.testOrg.ID).
		Pluck("name", &dbInstallNames).Error
	require.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), dbInstallNames, response.InstallNames,
		"response install_names should match database names")
}
