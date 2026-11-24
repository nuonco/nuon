package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installlookup "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/lookup"
)

func (s *InstallsServiceTestSuite) TestGetInstallByID() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), install.ID, resp.ID)
	assert.Equal(s.T(), install.Name, resp.Name)
	assert.Equal(s.T(), s.testApp.ID, resp.AppID)
}

func (s *InstallsServiceTestSuite) TestGetInstallByName() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s", install.Name)
	rr := s.makeRequest(http.MethodGet, path, nil)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), install.ID, resp.ID)
}

func (s *InstallsServiceTestSuite) TestGetInstallNotFound() {
	rr := s.makeRequest(http.MethodGet, "/v1/installs/ins_nonexistent_00000000", nil)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestGetInstallResolvesCloudPlatform() {
	for _, tc := range s.cloudPlatformResolutionTestCases() {
		s.Run(tc.name, func() {
			install := tc.setup()

			path := fmt.Sprintf("/v1/installs/%s", install.ID)
			rr := s.makeRequest(http.MethodGet, path, nil)
			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var resp app.Install
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
			assert.Equal(s.T(), tc.expectedCloudPlatform, resp.CloudPlatform)
			assert.Equal(s.T(), tc.expectedRunnerType, resp.RunnerType)
		})
	}
}

func (s *InstallsServiceTestSuite) TestInstallReferenceResolution() {
	shadow := s.createTestInstall()
	target := s.createTestInstall()
	if shadow.ID > target.ID {
		shadow, target = target, shadow
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: shadow.ID}).Update("name", target.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: target.ID}).Update("name", "prod").Error)

	siblingApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), siblingApp.ID)
	sibling := s.deps.Seeder.CreateInstall(s.ctx, s.T(), siblingApp)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: sibling.ID}).Update("name", "prod").Error)

	foreignCtx, _ := s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	foreignApp := s.deps.Seeder.CreateApp(foreignCtx, s.T())
	s.deps.Seeder.CreateAppConfig(foreignCtx, s.T(), foreignApp.ID)
	foreign := s.deps.Seeder.CreateInstall(foreignCtx, s.T(), foreignApp)
	require.NoError(s.T(), s.deps.DB.WithContext(foreignCtx).Model(&app.Install{ID: foreign.ID}).Update("name", "foreign-only").Error)
	unique := s.createTestInstall()

	for name, tc := range map[string]struct {
		ref    string
		wantID string
		status int
	}{
		"ID wins over lower-sorting name match": {target.ID, target.ID, http.StatusOK},
		"unique name":                           {unique.Name, unique.ID, http.StatusOK},
		"duplicate names across apps":           {"prod", "", http.StatusBadRequest},
		"foreign ID":                            {foreign.ID, "", http.StatusNotFound},
		"foreign name":                          {"foreign-only", "", http.StatusNotFound},
		"missing":                               {"missing-install", "", http.StatusNotFound},
	} {
		s.Run(name, func() {
			resolved, err := installlookup.ByRef(s.ctx, s.deps.DB.Preload("App"), s.testOrg.ID, tc.ref)
			if tc.status == http.StatusOK {
				require.NoError(s.T(), err)
				require.Equal(s.T(), tc.wantID, resolved.ID)
				require.Equal(s.T(), resolved.AppID, resolved.App.ID)
			} else {
				require.Nil(s.T(), resolved)
				if tc.status == http.StatusNotFound {
					require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
				} else {
					require.ErrorContains(s.T(), err, "ambiguous; use an install ID")
				}
			}

			rr := s.makeRequest(http.MethodGet, "/v1/installs/"+tc.ref, nil)
			require.Equal(s.T(), tc.status, rr.Code, "body: %s", rr.Body.String())
			result, _, err := s.installsService.mcpGetInstall(s.workflowMCPContext(), nil, mcpGetInstallInput{Install: tc.ref})
			if tc.status == http.StatusOK {
				require.NoError(s.T(), err)
				var response app.Install
				require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
				require.Equal(s.T(), tc.wantID, response.ID)
				require.NotNil(s.T(), result)
				require.Len(s.T(), result.Content, 1)
				text, ok := result.Content[0].(*mcp.TextContent)
				require.True(s.T(), ok)
				var overview mcpInstallOverview
				require.NoError(s.T(), json.Unmarshal([]byte(text.Text), &overview))
				require.Equal(s.T(), tc.wantID, overview.ID)
			} else if tc.status == http.StatusNotFound {
				require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
			} else {
				require.Contains(s.T(), rr.Body.String(), "ambiguous; use an install ID")
				require.ErrorContains(s.T(), err, "ambiguous; use an install ID")
			}
		})
	}

	for name, tc := range map[string]struct {
		appID  string
		ref    string
		wantID string
	}{
		"first app resolves prod":  {s.testApp.ID, "prod", target.ID},
		"second app resolves prod": {siblingApp.ID, "prod", sibling.ID},
		"scoped ID wins":           {s.testApp.ID, target.ID, target.ID},
		"ID cannot escape app":     {s.testApp.ID, sibling.ID, ""},
		"ID cannot escape org":     {foreignApp.ID, foreign.ID, ""},
		"name cannot escape org":   {foreignApp.ID, "foreign-only", ""},
	} {
		s.Run(name, func() {
			resolved, err := installlookup.OnApp(s.ctx, s.deps.DB.Preload("App"), s.testOrg.ID, tc.appID, tc.ref)
			if tc.wantID == "" {
				require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
				require.Nil(s.T(), resolved)
				return
			}
			require.NoError(s.T(), err)
			require.Equal(s.T(), tc.wantID, resolved.ID)
			require.Equal(s.T(), tc.appID, resolved.App.ID)
		})
	}

	require.NoError(s.T(), s.deps.DB.Delete(sibling).Error)
	resolved, err := installlookup.ByRef(s.ctx, s.deps.DB, s.testOrg.ID, "prod")
	require.NoError(s.T(), err)
	require.Equal(s.T(), target.ID, resolved.ID)
}
