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

func (s *ReleasesServiceTestSuite) TestGetRelease() {
	testCases := []struct {
		name         string
		setupFunc    func() string
		expectedCode int
		validateFunc func(*app.ComponentRelease)
	}{
		{
			name: "success with existing release",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "active",
					StatusDescription: "test release",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				return rel.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rel *app.ComponentRelease) {
				assert.NotEmpty(s.T(), rel.ID)
				assert.Equal(s.T(), "active", string(rel.Status))
				assert.Equal(s.T(), "test release", rel.StatusDescription)
				assert.NotEmpty(s.T(), rel.ComponentBuildID)
			},
		},
		{
			name: "success with release and steps",
			setupFunc: func() string {
				cmp := s.getSeededComponent(app.ComponentTypeTerraformModule)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "executing",
					StatusDescription: "test release with steps",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				// Create release steps
				step := &app.ComponentReleaseStep{
					ComponentReleaseID: rel.ID,
					Status:             "queued",
					StatusDescription:  "test step",
				}
				res = s.deps.DB.WithContext(s.ctx).Create(step)
				require.NoError(s.T(), res.Error)

				return rel.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(rel *app.ComponentRelease) {
				assert.NotEmpty(s.T(), rel.ID)
				assert.Equal(s.T(), "executing", string(rel.Status))
				assert.NotEmpty(s.T(), rel.ComponentReleaseSteps)
				assert.Equal(s.T(), 1, rel.TotalComponentReleaseSteps)
			},
		},
		{
			name: "error with non-existent release",
			setupFunc: func() string {
				return "rel_nonexistent12345678901"
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			releaseID := tc.setupFunc()
			path := fmt.Sprintf("/v1/releases/%s", releaseID)

			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var rel app.ComponentRelease
				err := json.Unmarshal(rr.Body.Bytes(), &rel)
				require.NoError(s.T(), err)

				if tc.validateFunc != nil {
					tc.validateFunc(&rel)
				}
			}
		})
	}
}

func (s *ReleasesServiceTestSuite) TestGetRelease_OrgIsolation() {
	s.Run("cannot see release from different org", func() {
		// Create release in original org
		cmp := s.getSeededComponent(app.ComponentTypeDockerBuild)
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

		path := fmt.Sprintf("/v1/releases/%s", rel.ID)
		req, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(s.T(), err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should not find the release (org isolation)
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
