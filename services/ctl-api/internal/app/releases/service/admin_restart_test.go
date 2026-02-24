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

func (s *ReleasesServiceTestSuite) TestRestartRelease() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		checkSignals bool
	}{
		{
			name: "success with existing release",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "error",
					StatusDescription: "test release in error state",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				return rel.ID
			},
			expectedCode: http.StatusOK,
			checkSignals: true,
		},
		{
			name: "error with non-existent release",
			setupFunc: func() string {
				return "rel_nonexistent12345678901"
			},
			expectedCode: http.StatusNotFound,
			checkSignals: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before each test case
			s.mockEvClient.Reset()

			releaseID := tc.setupFunc()
			path := fmt.Sprintf("/v1/releases/%s/admin-restart", releaseID)

			req := RestartReleaseReleaseRequest{}
			rr := s.makeRequest(http.MethodPost, path, req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var result bool
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				require.NoError(s.T(), err)
				assert.True(s.T(), result)

				// Verify event loop signal
				if tc.checkSignals {
					capturedSignals := s.mockEvClient.GetSignals()
					require.Len(s.T(), capturedSignals, 1)
					assert.Equal(s.T(), releaseID, capturedSignals[0].ID)

					sig, ok := capturedSignals[0].Signal.(*signals.Signal)
					require.True(s.T(), ok)
					assert.Equal(s.T(), signals.OperationRestart, sig.Type)
				}
			}

			if !tc.checkSignals {
				capturedSignals := s.mockEvClient.GetSignals()
				assert.Empty(s.T(), capturedSignals, "should not send signals on error")
			}
		})
	}
}

func (s *ReleasesServiceTestSuite) TestRestartRelease_OrgIsolation() {
	s.Run("cannot restart release from different org", func() {
		// Create release in original org
		cmp := s.getSeededComponent(app.ComponentTypeTerraformModule)
		ccc := s.getSeededConfigConnection(cmp.ID)
		build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		rel := &app.ComponentRelease{
			ComponentBuildID:  build.ID,
			Status:            "error",
			StatusDescription: "test release",
		}
		res := s.deps.DB.WithContext(s.ctx).Create(rel)
		require.NoError(s.T(), res.Error)

		// Create second org and try to restart
		ctx2, acc2 := s.deps.Seeder.EnsureAccount(s.ctx, s.T())
		_, org2 := s.deps.Seeder.EnsureOrg(ctx2, s.T())

		router := tests.NewTestRouter(tests.RouterOptions{
			L:       s.deps.L,
			DB:      s.deps.DB,
			TestOrg: org2,
			TestAcc: acc2,
		})
		err := s.releasesService.RegisterInternalRoutes(router)
		require.NoError(s.T(), err)

		path := fmt.Sprintf("/v1/releases/%s/admin-restart", rel.ID)

		reqBody := RestartReleaseReleaseRequest{}
		jsonBytes, err := json.Marshal(reqBody)
		require.NoError(s.T(), err)

		req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBytes))
		require.NoError(s.T(), err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should not find the release (org isolation)
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
