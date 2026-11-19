package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestGetAppFleetEmpty() {
	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/apps/%s/fleet", s.testApp.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp AppFleetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), int64(0), resp.Summary.Total)
	assert.Empty(s.T(), resp.Branches)
}

func (s *InstallsServiceTestSuite) TestGetAppFleetGroupsByBranchAndGroup() {
	branch := s.seedFleetBranch("fleet-groups", true)

	canary := s.createTestInstall()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.InstallAppBranchConnection{
		InstallID:      canary.ID,
		AppBranchID:    branch.ID,
		AppBranchGroup: "canary",
		Active:         true,
	}).Error)
	s.createTestInstall()

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/apps/%s/fleet", s.testApp.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp AppFleetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), int64(2), resp.Summary.Total)
	assert.Len(s.T(), resp.Branches, 2)

	var named *AppFleetBranch
	for _, b := range resp.Branches {
		if b.ID != nil && *b.ID == branch.ID {
			named = b
		}
	}
	require.NotNil(s.T(), named)
	require.NotNil(s.T(), named.TargetRun)
	require.Len(s.T(), named.Groups, 1)
	assert.Equal(s.T(), "canary", *named.Groups[0].Name)
	assert.Equal(s.T(), int64(1), named.Groups[0].Counts.Total)
	require.Len(s.T(), named.Groups[0].Installs, 1)
	assert.Equal(s.T(), canary.ID, named.Groups[0].Installs[0].ID)
}

func (s *InstallsServiceTestSuite) TestGetAppFleetRollout() {
	branch := s.seedFleetBranch("fleet-rollout", false)

	applied := s.createTestInstall()
	s.setFleetAppliedConfig(applied, s.testAppConfig.ID)
	s.connectFleetInstall(applied, branch)

	behind := s.createTestInstall()
	s.setFleetAppliedConfig(behind, "cfg-older-applied")
	s.connectFleetInstall(behind, branch)

	unknown := s.createTestInstall()
	s.connectFleetInstall(unknown, branch)

	rr := s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/apps/%s/fleet", s.testApp.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp AppFleetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), int64(1), resp.Summary.Applied)
	assert.Equal(s.T(), int64(1), resp.Summary.Behind)
	assert.Equal(s.T(), int64(1), resp.Summary.Unknown)

	rr = s.makeRequest(http.MethodGet, fmt.Sprintf("/v1/apps/%s/fleet?rollout=behind", s.testApp.ID), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var filtered AppFleetResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &filtered))
	assert.Equal(s.T(), int64(3), filtered.Summary.Total)
	assert.Equal(s.T(), int64(1), filtered.Summary.Matched)
	assert.Equal(s.T(), int64(1), filtered.Summary.Behind)
	assert.Equal(s.T(), int64(0), filtered.Summary.Applied)
	require.Len(s.T(), filtered.Branches, 1)
	require.Len(s.T(), filtered.Branches[0].Groups, 1)
	require.Len(s.T(), filtered.Branches[0].Groups[0].Installs, 1)
	assert.Equal(s.T(), behind.ID, filtered.Branches[0].Groups[0].Installs[0].ID)
}

func (s *InstallsServiceTestSuite) TestGetAppFleetInstallsPaginates() {
	branch := s.seedFleetBranch("fleet-page", false)
	for i := 0; i < 3; i++ {
		s.connectFleetInstall(s.createTestInstall(), branch)
	}

	path := fmt.Sprintf("/v1/apps/%s/fleet/installs?app_branch_id=%s&limit=2", s.testApp.ID, branch.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var first []AppFleetInstall
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &first))
	assert.Len(s.T(), first, 2)

	rr = s.makeRequest(http.MethodGet, path+"&offset=2", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var second []AppFleetInstall
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &second))
	assert.Len(s.T(), second, 1)
}

func (s *InstallsServiceTestSuite) TestGetAppFleetRejectsInvalidFilters() {
	for _, path := range []string{
		"/v1/apps/%s/fleet?health=bogus",
		"/v1/apps/%s/fleet?runner_status=sleeping",
		"/v1/apps/%s/fleet?rollout=ahead",
		"/v1/apps/%s/fleet?per_group=0",
	} {
		rr := s.makeRequest(http.MethodGet, fmt.Sprintf(path, s.testApp.ID), nil)
		assert.Equal(s.T(), http.StatusBadRequest, rr.Code, path)
	}
}

func (s *InstallsServiceTestSuite) seedFleetBranch(name string, withCanaryGroup bool) *app.AppBranch {
	branch := &app.AppBranch{
		AppID: s.testApp.ID,
		OrgID: s.testOrg.ID,
		Name:  fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	cfg := &app.AppBranchConfig{
		AppBranchID: branch.ID,
		OrgID:       s.testOrg.ID,
	}
	if withCanaryGroup {
		cfg.InstallGroups = []app.AppBranchInstallGroup{
			{OrgID: s.testOrg.ID, Name: "canary", Order: 1},
		}
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(cfg).Error)

	run := &app.AppBranchRun{
		OrgID:             s.testOrg.ID,
		AppBranchID:       branch.ID,
		AppBranchConfigID: cfg.ID,
		AppConfigID:       s.testAppConfig.ID,
		RunType:           app.AppBranchRunTypeGit,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(run).Error)
	return branch
}

func (s *InstallsServiceTestSuite) connectFleetInstall(install *app.Install, branch *app.AppBranch) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.InstallAppBranchConnection{
		InstallID:   install.ID,
		AppBranchID: branch.ID,
		Active:      true,
	}).Error)
}

func (s *InstallsServiceTestSuite) setFleetAppliedConfig(install *app.Install, appConfigID string) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{}).
		Where("id = ?", install.ID).
		Update("app_config_ref", app.AppConfigRef{AppliedConfigID: appConfigID}).Error)
}
