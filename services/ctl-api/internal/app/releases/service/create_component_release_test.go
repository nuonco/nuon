package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/releases/signals"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *ReleasesServiceTestSuite) TestCreateComponentRelease() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, CreateComponentReleaseRequest)
		expectedCode int
		validateFunc func(*app.ComponentRelease)
		checkSignals bool
	}{
		{
			name: "success with build_id and install_ids",
			setupFunc: func() (string, CreateComponentReleaseRequest) {
				cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
				install := s.createInstallForApp()

				req := CreateComponentReleaseRequest{
					BuildID:    build.ID,
					InstallIDs: []string{install.ID},
					Strategy: struct {
						InstallsPerStep int    `json:"installs_per_step"`
						Delay           string `json:"delay"`
					}{
						InstallsPerStep: 1,
						Delay:           "0s",
					},
				}
				return cmp.ID, req
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(rel *app.ComponentRelease) {
				assert.NotEmpty(s.T(), rel.ID)
				assert.Equal(s.T(), app.ReleaseStatus("queued"), rel.Status)
				assert.NotEmpty(s.T(), rel.ComponentBuildID)
			},
			checkSignals: true,
		},
		{
			name: "success with all installs",
			setupFunc: func() (string, CreateComponentReleaseRequest) {
				cmp := s.getSeededComponent(app.ComponentTypeTerraformModule)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
				s.createInstallForApp()
				s.createInstallForApp()

				req := CreateComponentReleaseRequest{
					BuildID: build.ID,
					Strategy: struct {
						InstallsPerStep int    `json:"installs_per_step"`
						Delay           string `json:"delay"`
					}{
						InstallsPerStep: 2,
						Delay:           "1m",
					},
				}
				return cmp.ID, req
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(rel *app.ComponentRelease) {
				assert.NotEmpty(s.T(), rel.ID)
				assert.Equal(s.T(), app.ReleaseStatus("queued"), rel.Status)
			},
			checkSignals: true,
		},
		{
			name: "error with missing build_id",
			setupFunc: func() (string, CreateComponentReleaseRequest) {
				cmp := s.getSeededComponent(app.ComponentTypeDockerBuild)
				s.createInstallForApp()

				req := CreateComponentReleaseRequest{
					Strategy: struct {
						InstallsPerStep int    `json:"installs_per_step"`
						Delay           string `json:"delay"`
					}{
						Delay: "0s",
					},
				}
				return cmp.ID, req
			},
			expectedCode: http.StatusBadRequest,
			checkSignals: false,
		},
		{
			name: "error with invalid delay",
			setupFunc: func() (string, CreateComponentReleaseRequest) {
				cmp := s.getSeededComponent(app.ComponentTypeKubernetesManifest)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
				s.createInstallForApp()

				req := CreateComponentReleaseRequest{
					BuildID: build.ID,
					Strategy: struct {
						InstallsPerStep int    `json:"installs_per_step"`
						Delay           string `json:"delay"`
					}{
						Delay: "invalid-delay",
					},
				}
				return cmp.ID, req
			},
			expectedCode: http.StatusBadRequest,
			checkSignals: false,
		},
		{
			name: "error with non-existent component",
			setupFunc: func() (string, CreateComponentReleaseRequest) {
				cmp := s.getSeededComponent(app.ComponentTypeExternalImage)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
				s.createInstallForApp()

				req := CreateComponentReleaseRequest{
					BuildID: build.ID,
					Strategy: struct {
						InstallsPerStep int    `json:"installs_per_step"`
						Delay           string `json:"delay"`
					}{
						Delay: "0s",
					},
				}
				return "cmp_nonexistent123456789012", req
			},
			expectedCode: http.StatusNotFound,
			checkSignals: false,
		},
		{
			name: "error with invalid install_id",
			setupFunc: func() (string, CreateComponentReleaseRequest) {
				cmp := s.getSeededComponent(app.ComponentTypeJob)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
				s.createInstallForApp()

				req := CreateComponentReleaseRequest{
					BuildID:    build.ID,
					InstallIDs: []string{"ins_nonexistent12345678901"},
					Strategy: struct {
						InstallsPerStep int    `json:"installs_per_step"`
						Delay           string `json:"delay"`
					}{
						Delay: "0s",
					},
				}
				return cmp.ID, req
			},
			expectedCode: http.StatusBadRequest,
			checkSignals: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before each test case
			s.mockEvClient.Reset()

			componentID, req := tc.setupFunc()
			path := fmt.Sprintf("/v1/components/%s/releases", componentID)

			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var rel app.ComponentRelease
				err := json.Unmarshal(rr.Body.Bytes(), &rel)
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(&rel)
				}

				// Verify database state
				var dbRelease app.ComponentRelease
				res := s.deps.DB.First(&dbRelease, "id = ?", rel.ID)
				require.NoError(s.T(), res.Error)
				assert.Equal(s.T(), rel.ComponentBuildID, dbRelease.ComponentBuildID)
				assert.Equal(s.T(), s.testOrg.ID, dbRelease.OrgID)

				// Verify event loop signals
				if tc.checkSignals {
					capturedSignals := s.mockEvClient.GetSignals()
					require.NotEmpty(s.T(), capturedSignals, "expected event loop signals to be sent")

					// Should send 3 signals: OperationCreated, OperationPollDependencies, OperationProvision
					assert.GreaterOrEqual(s.T(), len(capturedSignals), 3)

					foundCreated := false
					foundPollDeps := false
					foundProvision := false

					for _, sig := range capturedSignals {
						if relSig, ok := sig.Signal.(*signals.Signal); ok {
							switch relSig.Type {
							case signals.OperationCreated:
								foundCreated = true
							case signals.OperationPollDependencies:
								foundPollDeps = true
							case signals.OperationProvision:
								foundProvision = true
							}
						}
					}

					assert.True(s.T(), foundCreated, "expected OperationCreated signal")
					assert.True(s.T(), foundPollDeps, "expected OperationPollDependencies signal")
					assert.True(s.T(), foundProvision, "expected OperationProvision signal")
				}
			}

			if !tc.checkSignals {
				capturedSignals := s.mockEvClient.GetSignals()
				assert.Empty(s.T(), capturedSignals, "should not send signals on error")
			}
		})
	}
}

func (s *ReleasesServiceTestSuite) TestCreateComponentRelease_OrgIsolation() {
	s.Run("cannot create release for component in different org", func() {
		// Create second org
		ctx2, acc2 := s.deps.Seeder.EnsureAccount(s.ctx, s.T())
		_, org2 := s.deps.Seeder.EnsureOrg(ctx2, s.T())

		// Create component in original org
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)
		build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
		s.createInstallForApp()

		// Try to create release using second org's router
		router := tests.NewTestRouter(tests.RouterOptions{
			L:       s.deps.L,
			DB:      s.deps.DB,
			TestOrg: org2,
			TestAcc: acc2,
		})
		err := s.releasesService.RegisterPublicRoutes(router)
		require.NoError(s.T(), err)

		req := CreateComponentReleaseRequest{
			BuildID: build.ID,
			Strategy: struct {
				InstallsPerStep int    `json:"installs_per_step"`
				Delay           string `json:"delay"`
			}{
				Delay: "0s",
			},
		}

		path := fmt.Sprintf("/v1/components/%s/releases", cmp.ID)

		var reqBody *bytes.Buffer
		jsonBytes, err := json.Marshal(req)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)

		httpReq, err := http.NewRequest(http.MethodPost, path, reqBody)
		require.NoError(s.T(), err)
		httpReq.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httpReq)

		// Should not find the component (org isolation)
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
