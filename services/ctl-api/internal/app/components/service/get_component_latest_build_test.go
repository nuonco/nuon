package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentLatestBuildSuccess() {
	s.Run("returns latest build", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)
		seededBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/latest", s.testApp.ID, cmp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.ComponentBuild
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), seededBuild.ID, response.ID)
	})
}

func (s *ComponentsServiceTestSuite) TestGetAppComponentLatestBuildExcludesPreviewBuilds() {
	tests := map[string]struct {
		runType       app.AppBranchRunType
		planOnly      bool
		createPreview bool
	}{
		"preview child": {
			runType:       app.AppBranchRunTypeManual,
			createPreview: true,
		},
		"legacy preview run": {
			runType: app.AppBranchRunTypeGitPreview,
		},
		"legacy plan only run": {
			runType:  app.AppBranchRunTypeManual,
			planOnly: true,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
			ccc := s.getSeededConfigConnection(cmp.ID)
			expected := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
			previewBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

			branch := app.AppBranch{AppID: s.testApp.ID, Name: "preview-build-filter-" + name}
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&branch).Error)
			branchConfig := app.AppBranchConfig{AppBranchID: branch.ID}
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&branchConfig).Error)
			run := app.AppBranchRun{
				AppBranchID:       branch.ID,
				AppBranchConfigID: branchConfig.ID,
				AppConfigID:       s.testAppConfig.ID,
				RunType:           test.runType,
				PlanOnly:          test.planOnly,
			}
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&run).Error)
			if test.createPreview {
				preview := app.AppBranchRunPreview{
					AppBranchRunID: run.ID,
					Source:         app.AppBranchRunPreviewSourceBranch,
					Mode:           app.AppBranchRunPreviewModeBuildOnly,
				}
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&preview).Error)
			}
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
				Model(previewBuild).
				Update("app_branch_run_id", run.ID).
				Error)

			path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/latest", s.testApp.ID, cmp.ID)
			rr := s.makeRequest(http.MethodGet, path, nil)
			require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

			var response app.ComponentBuild
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
			assert.Equal(s.T(), expected.ID, response.ID)
		})
	}
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentLatestBuildNotFound() {
	s.Run("no builds exist for component", func() {
		// Create a fresh component with a config connection but no builds
		freshComp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)
		s.deps.Seeder.CreateDockerBuildComponentConfigConnection(s.ctx, s.T(), freshComp.ID, s.testAppConfig.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/latest", s.testApp.ID, freshComp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
