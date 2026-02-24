package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *ReleasesServiceTestSuite) TestGetComponentReleases() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		expectedCode  int
		expectedCount int
		validateFunc  func([]app.ComponentRelease)
	}{
		{
			name: "success with releases",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build1 := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
				build2 := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				// Create releases
				rel1 := &app.ComponentRelease{
					ComponentBuildID:  build1.ID,
					Status:            "queued",
					StatusDescription: "test release 1",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel1)
				require.NoError(s.T(), res.Error)

				rel2 := &app.ComponentRelease{
					ComponentBuildID:  build2.ID,
					Status:            "active",
					StatusDescription: "test release 2",
				}
				res = s.deps.DB.WithContext(s.ctx).Create(rel2)
				require.NoError(s.T(), res.Error)

				return cmp.ID
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 2,
			validateFunc: func(releases []app.ComponentRelease) {
				assert.Len(s.T(), releases, 2)
			},
		},
		{
			name: "success with pagination",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeTerraformModule)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				// Create multiple releases
				for i := 0; i < 5; i++ {
					rel := &app.ComponentRelease{
						ComponentBuildID:  build.ID,
						Status:            "queued",
						StatusDescription: fmt.Sprintf("test release %d", i),
					}
					res := s.deps.DB.WithContext(s.ctx).Create(rel)
					require.NoError(s.T(), res.Error)
				}

				return cmp.ID
			},
			queryParams:   "?limit=2",
			expectedCode:  http.StatusOK,
			expectedCount: 2,
		},
		{
			name: "success with empty results",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeDockerBuild)
				return cmp.ID
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
		{
			name: "error with non-existent component",
			setupFunc: func() string {
				return "cmp_nonexistent123456789012"
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			componentID := tc.setupFunc()
			path := fmt.Sprintf("/v1/components/%s/releases%s", componentID, tc.queryParams)

			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var releases []app.ComponentRelease
				err := json.Unmarshal(rr.Body.Bytes(), &releases)
				require.NoError(s.T(), err)

				assert.Len(s.T(), releases, tc.expectedCount)

				if tc.validateFunc != nil {
					tc.validateFunc(releases)
				}
			}
		})
	}
}

func (s *ReleasesServiceTestSuite) TestGetComponentReleases_OrgIsolation() {
	s.Run("cannot see releases from different org", func() {
		// Create releases in original org
		cmp := s.getSeededComponent(app.ComponentTypeKubernetesManifest)
		ccc := s.getSeededConfigConnection(cmp.ID)
		build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		rel := &app.ComponentRelease{
			ComponentBuildID:  build.ID,
			Status:            "queued",
			StatusDescription: "test release",
		}
		res := s.deps.DB.WithContext(s.ctx).Create(rel)
		require.NoError(s.T(), res.Error)

		// Create second org and try to access
		ctx2, acc2 := s.deps.Seeder.EnsureAccount(s.ctx, s.T())
		_, org2 := s.deps.Seeder.EnsureOrg(ctx2, s.T())

		router := tests.NewTestRouter(tests.RouterOptions{
			L:       s.deps.L,
			DB:      s.deps.DB,
			TestOrg: org2,
			TestAcc: acc2,
		})
		err := s.releasesService.RegisterPublicRoutes(router)
		require.NoError(s.T(), err)

		path := fmt.Sprintf("/v1/components/%s/releases", cmp.ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(s.T(), err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var releases []app.ComponentRelease
		err = json.Unmarshal(rr.Body.Bytes(), &releases)
		require.NoError(s.T(), err)

		// Should return empty array (org isolation)
		assert.Empty(s.T(), releases)
	})
}
