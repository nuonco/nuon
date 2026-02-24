package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *ReleasesServiceTestSuite) TestGetAllReleases() {
	testCases := []struct {
		name         string
		setupFunc    func()
		queryParams  string
		expectedCode int
		minCount     int
		validateFunc func([]*app.ComponentRelease)
	}{
		{
			name: "success with releases from multiple orgs",
			setupFunc: func() {
				// Create releases in test org
				cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
				ccc := s.getSeededConfigConnection(cmp.ID)
				build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

				rel := &app.ComponentRelease{
					ComponentBuildID:  build.ID,
					Status:            "queued",
					StatusDescription: "test release org 1",
				}
				res := s.deps.DB.WithContext(s.ctx).Create(rel)
				require.NoError(s.T(), res.Error)

				// Create release in second org
				ctx2, _ := s.deps.Seeder.EnsureAccount(s.ctx, s.T())
				ctx2, _ = s.deps.Seeder.EnsureOrg(ctx2, s.T())
				app2 := s.deps.Seeder.CreateApp(ctx2, s.T())
				cfg2 := s.deps.Seeder.CreateAppConfig(ctx2, s.T(), app2.ID)

				var cmp2 app.Component
				res = s.deps.DB.WithContext(ctx2).First(&cmp2, "app_id = ?", app2.ID)
				require.NoError(s.T(), res.Error)

				var ccc2 app.ComponentConfigConnection
				res = s.deps.DB.WithContext(ctx2).First(&ccc2, "component_id = ? AND app_config_id = ?", cmp2.ID, cfg2.ID)
				require.NoError(s.T(), res.Error)

				build2 := s.deps.Seeder.CreateComponentBuild(ctx2, s.T(), ccc2.ID)

				rel2 := &app.ComponentRelease{
					ComponentBuildID:  build2.ID,
					Status:            "active",
					StatusDescription: "test release org 2",
				}
				res = s.deps.DB.WithContext(ctx2).Create(rel2)
				require.NoError(s.T(), res.Error)
			},
			queryParams:  "",
			expectedCode: http.StatusOK,
			minCount:     2,
			validateFunc: func(releases []*app.ComponentRelease) {
				// Should return releases from all orgs (no org filtering)
				assert.GreaterOrEqual(s.T(), len(releases), 2)
			},
		},
		{
			name: "success with pagination",
			setupFunc: func() {
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
			},
			queryParams:  "?limit=3",
			expectedCode: http.StatusOK,
			minCount:     3,
		},
		{
			name:         "success with empty results",
			setupFunc:    func() {},
			queryParams:  "",
			expectedCode: http.StatusOK,
			minCount:     0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupFunc()
			path := fmt.Sprintf("/v1/releases%s", tc.queryParams)

			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var releases []*app.ComponentRelease
				err := json.Unmarshal(rr.Body.Bytes(), &releases)
				require.NoError(s.T(), err)

				assert.GreaterOrEqual(s.T(), len(releases), tc.minCount)

				if tc.validateFunc != nil {
					tc.validateFunc(releases)
				}
			}
		})
	}
}
