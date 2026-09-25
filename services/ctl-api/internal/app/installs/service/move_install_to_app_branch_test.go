package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appbranchchanged"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type branchWithStaleRun struct {
	branch            *app.AppBranch
	defaultGroupName  string
	pinnedGroupName   string
	runDefaultGroupID string
}

func (s *InstallsServiceTestSuite) seedBranchWithPinnedGroupAfterRun() branchWithStaleRun {
	branch := &app.AppBranch{
		AppID: s.testApp.ID,
		OrgID: s.testOrg.ID,
		Name:  fmt.Sprintf("stale-run-%d", time.Now().UnixNano()),
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	v1 := &app.AppBranchConfig{
		AppBranchID: branch.ID,
		CreatedAt:   time.Now().UTC().Add(-time.Minute),
		InstallGroups: []app.AppBranchInstallGroup{
			{Name: "Default", Order: 0, Default: true},
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(v1).Error)

	run := &app.AppBranchRun{
		AppBranchID:       branch.ID,
		AppBranchConfigID: v1.ID,
		AppConfigID:       s.testAppConfig.ID,
		RunType:           app.AppBranchRunTypeGit,
		PlanOnly:          false,
		Labeled: labels.Labeled{
			Labels: labels.Labels{app.AppBranchRunLabelBuildsCompleted: "true"},
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(run).Error)

	v2 := &app.AppBranchConfig{
		AppBranchID: branch.ID,
		CreatedAt:   time.Now().UTC(),
		InstallGroups: []app.AppBranchInstallGroup{
			{Name: "Default", Order: 0, Default: true},
			{Name: "manualabc", Order: 1},
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(v2).Error)

	var runDefault app.AppBranchInstallGroup
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.AppBranchInstallGroup{AppBranchConfigID: v1.ID, Name: "Default"}).
		First(&runDefault).Error)

	return branchWithStaleRun{
		branch:            branch,
		defaultGroupName:  "Default",
		pinnedGroupName:   "manualabc",
		runDefaultGroupID: runDefault.ID,
	}
}

func (s *InstallsServiceTestSuite) createInstallWithSignalsQueue() *app.Install {
	install := s.createTestInstall()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.Queue{
		OwnerID:   install.ID,
		OwnerType: "installs",
		Name:      helpers.InstallSignalsQueueName,
	}).Error)
	return install
}

func (s *InstallsServiceTestSuite) activeBranchConnection(installID, branchID string) app.InstallAppBranchConnection {
	var connection app.InstallAppBranchConnection
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallAppBranchConnection{
			InstallID:   installID,
			AppBranchID: branchID,
			Active:      true,
		}).
		First(&connection).Error)
	return connection
}

func (s *InstallsServiceTestSuite) lastAppBranchChangedSignal(installID string) appbranchchanged.Signal {
	signals := tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, installID)
	var payload appbranchchanged.Signal
	found := false
	for _, queued := range signals {
		if queued.Type != appbranchchanged.SignalType {
			continue
		}
		raw, err := json.Marshal(queued.Signal)
		require.NoError(s.T(), err)
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		require.NoError(s.T(), json.Unmarshal(raw, &envelope))
		require.NoError(s.T(), json.Unmarshal(envelope.Data, &payload))
		found = true
	}
	require.True(s.T(), found, "expected an app-branch-changed signal")
	return payload
}

func (s *InstallsServiceTestSuite) TestMoveInstallToAppBranchPinsGroupAddedAfterLastRun() {
	fixture := s.seedBranchWithPinnedGroupAfterRun()
	install := s.createInstallWithSignalsQueue()

	rr := s.makeRequest(http.MethodPatch, fmt.Sprintf("/v1/installs/%s/app-branch", install.ID), MoveInstallToAppBranchRequest{
		AppBranchID:    fixture.branch.ID,
		AppBranchGroup: fixture.pinnedGroupName,
	})
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	connection := s.activeBranchConnection(install.ID, fixture.branch.ID)
	assert.Equal(s.T(), fixture.pinnedGroupName, connection.AppBranchGroup)
	assert.Equal(s.T(), app.InstallAppBranchGroupAssignmentSourceExplicit, connection.AppBranchGroupAssignmentSource)

	payload := s.lastAppBranchChangedSignal(install.ID)
	assert.Empty(s.T(), payload.InstallGroupID)
}

func (s *InstallsServiceTestSuite) TestMoveInstallToAppBranchRecordsRunGroupIDWhenNameExistsOnRun() {
	fixture := s.seedBranchWithPinnedGroupAfterRun()
	install := s.createInstallWithSignalsQueue()

	rr := s.makeRequest(http.MethodPatch, fmt.Sprintf("/v1/installs/%s/app-branch", install.ID), MoveInstallToAppBranchRequest{
		AppBranchID:    fixture.branch.ID,
		AppBranchGroup: fixture.defaultGroupName,
	})
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	connection := s.activeBranchConnection(install.ID, fixture.branch.ID)
	assert.Equal(s.T(), fixture.defaultGroupName, connection.AppBranchGroup)
	assert.Equal(s.T(), app.InstallAppBranchGroupAssignmentSourceExplicit, connection.AppBranchGroupAssignmentSource)

	payload := s.lastAppBranchChangedSignal(install.ID)
	assert.Equal(s.T(), fixture.runDefaultGroupID, payload.InstallGroupID)
}

func (s *InstallsServiceTestSuite) TestMoveInstallToAppBranchRejectsUnknownGroup() {
	fixture := s.seedBranchWithPinnedGroupAfterRun()
	install := s.createTestInstall()

	rr := s.makeRequest(http.MethodPatch, fmt.Sprintf("/v1/installs/%s/app-branch", install.ID), MoveInstallToAppBranchRequest{
		AppBranchID:    fixture.branch.ID,
		AppBranchGroup: "does-not-exist",
	})
	if rr.Code != http.StatusBadRequest {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)

	var resp stderr.ErrResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(s.T(), resp.Error, "selects unknown app branch group does-not-exist")
}
