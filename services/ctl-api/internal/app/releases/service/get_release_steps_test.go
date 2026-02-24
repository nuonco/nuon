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

func (s *ReleasesServiceTestSuite) TestGetReleaseSteps() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		expectedCode  int
		expectedCount int
		validateFunc  func([]app.ComponentReleaseStep)
	}{
		{
			name: "success with multiple steps",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "executing",
					StatusDescription: "test release",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				// Create multiple steps
				for i := 0; i < 3; i++ {
					step := &app.ComponentReleaseStep{
						ComponentReleaseID: rel.ID,
						Status:             "queued",
						StatusDescription:  fmt.Sprintf("test step %d", i),
					}
					res = s.deps.DB.WithContext(s.ctx).Create(step)
					require.NoError(s.T(), res.Error)
				}

				return rel.ID
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(steps []app.ComponentReleaseStep) {
				assert.Len(s.T(), steps, 3)
				for _, step := range steps {
					assert.NotEmpty(s.T(), step.ID)
					assert.Equal(s.T(), "queued", step.Status)
				}
			},
		},
		{
			name: "success with pagination",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeTerraformModule)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "executing",
					StatusDescription: "test release",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				// Create multiple steps
				for i := 0; i < 5; i++ {
					step := &app.ComponentReleaseStep{
						ComponentReleaseID: rel.ID,
						Status:             "queued",
						StatusDescription:  fmt.Sprintf("test step %d", i),
					}
					res = s.deps.DB.WithContext(s.ctx).Create(step)
					require.NoError(s.T(), res.Error)
				}

				return rel.ID
			},
			queryParams:   "?limit=2",
			expectedCode:  http.StatusOK,
			expectedCount: 2,
		},
		{
			name: "success with empty results",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeDockerBuild)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "queued",
					StatusDescription: "test release with no steps",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				return rel.ID
			},
			queryParams:   "",
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
		{
			name: "error with non-existent release",
			setupFunc: func() string {
				return "rel_nonexistent12345678901"
			},
			queryParams:  "",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			releaseID := tc.setupFunc()
			path := fmt.Sprintf("/v1/releases/%s/steps%s", releaseID, tc.queryParams)

			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var steps []app.ComponentReleaseStep
				err := json.Unmarshal(rr.Body.Bytes(), &steps)
				require.NoError(s.T(), err)

				assert.Len(s.T(), steps, tc.expectedCount)

				if tc.validateFunc != nil {
					tc.validateFunc(steps)
				}
			}
		})
	}
}

func (s *ReleasesServiceTestSuite) TestGetReleaseSteps_OrgIsolation() {
	s.Run("cannot see steps from release in different org", func() {
		// Create release in original org
		cmp := s.getSeededComponent(app.ComponentTypeKubernetesManifest)
		ccc := s.getSeededConfigConnection(cmp.ID)
		build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		rel := &app.ComponentRelease{
			ComponentBuildID:  build.ID,
			Status:            "executing",
			StatusDescription: "test release",
		}
		res := s.deps.DB.WithContext(s.ctx).Create(rel)
		require.NoError(s.T(), res.Error)

		step := &app.ComponentReleaseStep{
			ComponentReleaseID: rel.ID,
			Status:             "queued",
			StatusDescription:  "test step",
		}
		res = s.deps.DB.WithContext(s.ctx).Create(step)
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

		path := fmt.Sprintf("/v1/releases/%s/steps", rel.ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(s.T(), err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should not find the release (org isolation)
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
