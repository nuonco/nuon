package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

var branchLabelsTestSeq int64

// uniqueBranchName keeps branch names collision-free across the many s.Run
// subtests sharing one suite-level app, since app branch names are unique per app.
func uniqueBranchName(name string) string {
	return fmt.Sprintf("%s-%d", name, atomic.AddInt64(&branchLabelsTestSeq, 1))
}

type branchCfgState int

const (
	branchCfgStateNone branchCfgState = iota
	branchCfgStateActive
	branchCfgStateBuilding
	branchCfgStatePreviewOnly
)

// seedBranch creates an app branch, optionally with a config in cfgState and the
// given install groups on that config. Returns the branch and its config (nil
// when cfgState is branchCfgStateNone).
func (s *InstallsServiceTestSuite) seedBranch(name string, groups []app.AppBranchInstallGroup, cfgState branchCfgState) (*app.AppBranch, *app.AppConfig) {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: uniqueBranchName(name)}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	if cfgState == branchCfgStateNone {
		return branch, nil
	}

	cfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	updates := map[string]any{"app_branch_id": branch.ID}
	switch cfgState {
	case branchCfgStateActive:
		updates["builds_pinned"] = true
		updates["builds_locked"] = true
	case branchCfgStateBuilding:
		updates["builds_pinned"] = false
		updates["builds_locked"] = false
	case branchCfgStatePreviewOnly:
		updates["builds_pinned"] = true
		updates["builds_locked"] = true
		updates["labels"] = labels.Labels{"source": string(app.AppBranchRunTypeGitPreview)}
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(cfg).Updates(updates).Error)

	for i := range groups {
		groups[i].Order = i
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
		AppBranchID:   branch.ID,
		InstallGroups: groups,
	}).Error)

	return branch, cfg
}

// removeFallbackAppConfig soft-deletes the suite's default non-branch app config, so
// the app has no latest apps-sync config for an unbranched or unmatched install to fall back to.
func (s *InstallsServiceTestSuite) removeFallbackAppConfig() {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(s.testAppConfig).Error)
}

func (s *InstallsServiceTestSuite) createInstallWithLabels(name string, reqLabels map[string]string) *httpRecorderResult {
	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:       name,
			AWSAccount: &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
			Labels:     reqLabels,
		},
	}
	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	return &httpRecorderResult{code: rr.Code, body: rr.Body.Bytes()}
}

type httpRecorderResult struct {
	code int
	body []byte
}

func (r *httpRecorderResult) install(t require.TestingT) *app.Install {
	var install app.Install
	require.NoError(t, json.Unmarshal(r.body, &install))
	return &install
}

func (s *InstallsServiceTestSuite) assertActiveConnection(installID, branchID, group string, source app.InstallAppBranchGroupAssignmentSource) {
	var connection app.InstallAppBranchConnection
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallAppBranchConnection{InstallID: installID, AppBranchID: branchID, Active: true}).
		First(&connection).Error)
	assert.Equal(s.T(), group, connection.AppBranchGroup)
	assert.Equal(s.T(), source, connection.AppBranchGroupAssignmentSource)
}

func (s *InstallsServiceTestSuite) assertNoInstallCreated(name string) {
	var count int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{}).
		Where(app.Install{AppID: s.testApp.ID, Name: name}).
		Count(&count).Error)
	assert.Zero(s.T(), count, "expected no install to be created")
}

func defaultGroup() app.AppBranchInstallGroup {
	return app.AppBranchInstallGroup{Name: "default", Default: true}
}

func selectingGroup(name string, match labels.Labels) app.AppBranchInstallGroup {
	return app.AppBranchInstallGroup{Name: name, LabelSelector: &labels.Selector{MatchLabels: match}}
}

func (s *InstallsServiceTestSuite) TestCreateInstallChoosesAmongAppBranchesByLabels() {
	s.Run("no branches and no labels uses the fallback config", func() {
		s.expectQueueCreation()
		res := s.createInstallWithLabels("no-branches-no-labels", nil)
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), s.testAppConfig.ID, install.AppConfigID)
	})

	s.Run("no branches with labels still uses the fallback config", func() {
		s.expectQueueCreation()
		res := s.createInstallWithLabels("no-branches-with-labels", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), s.testAppConfig.ID, install.AppConfigID)
	})

	s.Run("branch with unlabeled default group and no request labels does not join", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{defaultGroup()}, branchCfgStateActive)
		res := s.createInstallWithLabels("unlabeled-default-no-labels", nil)
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
	})

	s.Run("branch with unlabeled default group and labeled request does not join", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{defaultGroup()}, branchCfgStateActive)
		res := s.createInstallWithLabels("unlabeled-default-with-labels", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
	})

	s.Run("branch with a group selecting matching labels joins that branch", func() {
		s.expectQueueCreation()
		branch, cfg := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("matching-labels-joins", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), branch.ID, install.AppBranchID.String)
		assert.Equal(s.T(), cfg.ID, install.AppConfigID)
	})

	s.Run("branch selecting env=prod falls back when labels do not match", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("non-matching-labels-fallback", map[string]string{"env": "stage"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), s.testAppConfig.ID, install.AppConfigID)
	})

	s.Run("branch selecting env=prod falls back when request has no labels", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("no-labels-fallback", nil)
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
	})

	s.Run("labels selecting only one of two branches joins it", func() {
		s.expectQueueCreation()
		branchA, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		s.seedBranch("stage", []app.AppBranchInstallGroup{selectingGroup("stage", labels.Labels{"env": "stage"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("two-branches-one-matches", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), branchA.ID, install.AppBranchID.String)
	})

	s.Run("labels matching two branches returns a 400 naming both", func() {
		branchA, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		branchB, _ := s.seedBranch("acme", []app.AppBranchInstallGroup{selectingGroup("acme", labels.Labels{"customer": "acme"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("two-branches-both-match", map[string]string{"env": "prod", "customer": "acme"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), branchA.Name)
		assert.Contains(s.T(), string(res.body), branchB.Name)
		s.assertNoInstallCreated("two-branches-both-match")
	})

	s.Run("two branches selecting the same labels returns a 400 naming both", func() {
		branchA, _ := s.seedBranch("prod-a", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		branchB, _ := s.seedBranch("prod-b", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("two-branches-same-selector", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), branchA.Name)
		assert.Contains(s.T(), string(res.body), branchB.Name)
	})

	s.Run("matching branch with an unlabeled default sibling still joins the matching branch", func() {
		s.expectQueueCreation()
		branchA, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		s.seedBranch("other", []app.AppBranchInstallGroup{defaultGroup()}, branchCfgStateActive)
		res := s.createInstallWithLabels("matching-with-default-sibling", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), branchA.ID, install.AppBranchID.String)
	})

	s.Run("soft-deleted matching branch is ignored, falls back", func() {
		s.expectQueueCreation()
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(branch).Error)
		res := s.createInstallWithLabels("soft-deleted-branch-fallback", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
	})

	s.Run("another app's matching branch is ignored, falls back", func() {
		s.expectQueueCreation()
		otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
		otherBranch := &app.AppBranch{AppID: otherApp.ID, Name: "prod"}
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(otherBranch).Error)
		cfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), otherApp.ID)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(cfg).
			Updates(map[string]any{"app_branch_id": otherBranch.ID, "builds_pinned": true, "builds_locked": true}).Error)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
			AppBranchID: otherBranch.ID,
			InstallGroups: []app.AppBranchInstallGroup{{
				Name:          "prod",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
			}},
		}).Error)

		res := s.createInstallWithLabels("other-app-branch-fallback", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
	})

	s.Run("only the branch's latest config selectors are considered", func() {
		s.expectQueueCreation()
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		// A newer config on the same branch now selects a different label, so the
		// older config's selector must not be consulted.
		newerCfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(newerCfg).
			Updates(map[string]any{"app_branch_id": branch.ID, "builds_pinned": true, "builds_locked": true}).Error)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
			AppBranchID: branch.ID,
			InstallGroups: []app.AppBranchInstallGroup{{
				Name:          "stage",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "stage"}},
			}},
		}).Error)

		res := s.createInstallWithLabels("older-config-selector-stale", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
	})
}

func (s *InstallsServiceTestSuite) TestCreateInstallLabelsUsedForBranchMatching() {
	s.Run("app default labels alone match a branch", func() {
		s.expectQueueCreation()
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(s.testApp).
			Update("default_labels", labels.Labels{"env": "prod"}).Error)
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)

		res := s.createInstallWithLabels("default-labels-match", nil)
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), branch.ID, install.AppBranchID.String)
	})

	s.Run("request label conflicting with app default label still errors before matching", func() {
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(s.testApp).
			Update("default_labels", labels.Labels{"env": "prod"}).Error)
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)

		res := s.createInstallWithLabels("conflicting-default-label", map[string]string{"env": "stage"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "cannot be overridden")
	})

	s.Run("only templated labels do not match a branch and the template is stored", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "*"})}, branchCfgStateActive)

		res := s.createInstallWithLabels("templated-label-no-match", map[string]string{"env": "{{ .nuon.install.name }}"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.False(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), "{{ .nuon.install.name }}", install.LabelTemplates["env"])
	})

	s.Run("a static label alongside a templated one still matches a branch", func() {
		s.expectQueueCreation()
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)

		res := s.createInstallWithLabels("static-and-templated-labels", map[string]string{
			"env":  "prod",
			"name": "{{ .nuon.install.name }}",
		})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), branch.ID, install.AppBranchID.String)
	})
}

func (s *InstallsServiceTestSuite) TestCreateInstallGroupInsideMatchedBranch() {
	s.Run("joining by labels assigns the matching labeled group", func() {
		s.expectQueueCreation()
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		res := s.createInstallWithLabels("joins-labeled-group", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		s.assertActiveConnection(install.ID, branch.ID, "prod", app.InstallAppBranchGroupAssignmentSourceLabels)
	})

	s.Run("joining by labels assigns a default group that is itself selected by labels", func() {
		s.expectQueueCreation()
		group := defaultGroup()
		group.LabelSelector = &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{group}, branchCfgStateActive)
		res := s.createInstallWithLabels("joins-labeled-default-group", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		s.assertActiveConnection(install.ID, branch.ID, "default", app.InstallAppBranchGroupAssignmentSourceLabels)
	})

	s.Run("labels matching two groups on the matched branch error without creating an install", func() {
		s.seedBranch("prod", []app.AppBranchInstallGroup{
			selectingGroup("gold", labels.Labels{"env": "prod", "tier": "gold"}),
			selectingGroup("prod", labels.Labels{"env": "prod"}),
		}, branchCfgStateActive)
		res := s.createInstallWithLabels("matches-two-groups", map[string]string{"env": "prod", "tier": "gold"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "more than one install group")
		s.assertNoInstallCreated("matches-two-groups")
	})

	s.Run("explicit app_branch_group on a label-matched branch wins over label matching", func() {
		s.expectQueueCreation()
		branch, _ := s.seedBranch("prod", []app.AppBranchInstallGroup{
			selectingGroup("gold", labels.Labels{"env": "prod"}),
			{Name: "silver"},
		}, branchCfgStateActive)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:           "explicit-group-on-matched-branch",
				AWSAccount:     &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				Labels:         map[string]string{"env": "prod"},
				AppBranchGroup: "silver",
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
		var install app.Install
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
		s.assertActiveConnection(install.ID, branch.ID, "silver", app.InstallAppBranchGroupAssignmentSourceExplicit)
	})

	s.Run("explicit app_branch_group naming an unknown group on a label-matched branch errors", func() {
		s.seedBranch("prod", []app.AppBranchInstallGroup{
			selectingGroup("gold", labels.Labels{"env": "prod"}),
		}, branchCfgStateActive)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:           "unknown-group-on-matched-branch",
				AWSAccount:     &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				Labels:         map[string]string{"env": "prod"},
				AppBranchGroup: "missing",
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
		assert.Contains(s.T(), rr.Body.String(), "does not exist")
	})

	s.Run("app_branch_group is ignored when no branch matches by labels", func() {
		s.expectQueueCreation()
		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:           "group-ignored-no-match",
				AWSAccount:     &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				AppBranchGroup: "prod",
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
		var install app.Install
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
		assert.False(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), s.testAppConfig.ID, install.AppConfigID)
	})
}

func (s *InstallsServiceTestSuite) TestCreateInstallExplicitAppBranchUnchangedByLabelMatching() {
	s.Run("explicit app_branch wins even when labels match a different branch", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		branchB, _ := s.seedBranch("stage", []app.AppBranchInstallGroup{defaultGroup()}, branchCfgStateActive)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:        "explicit-branch-wins",
				AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				Labels:      map[string]string{"env": "prod"},
				AppBranchID: branchB.ID,
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
		var install app.Install
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), branchB.ID, install.AppBranchID.String)
	})

	s.Run("explicit app_branch with a default group and non-matching labels uses the default group", func() {
		s.expectQueueCreation()
		branch, _ := s.seedBranch("stage", []app.AppBranchInstallGroup{defaultGroup()}, branchCfgStateActive)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:        "explicit-branch-default-group",
				AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				Labels:      map[string]string{"env": "nothing-matches"},
				AppBranchID: branch.ID,
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
		var install app.Install
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
		s.assertActiveConnection(install.ID, branch.ID, "default", app.InstallAppBranchGroupAssignmentSourceDefault)
	})

	s.Run("explicit app_branch with no default group and non-matching labels errors", func() {
		branch, _ := s.seedBranch("stage", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:        "explicit-branch-no-default-group",
				AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				AppBranchID: branch.ID,
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
		assert.Contains(s.T(), rr.Body.String(), "does not match an install group")
	})

	s.Run("explicit app_branch avoids the ambiguous-match error even when labels fit two other branches", func() {
		s.expectQueueCreation()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		s.seedBranch("acme", []app.AppBranchInstallGroup{selectingGroup("acme", labels.Labels{"customer": "acme"})}, branchCfgStateActive)
		explicit, _ := s.seedBranch("target", []app.AppBranchInstallGroup{defaultGroup()}, branchCfgStateActive)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:        "explicit-branch-avoids-ambiguity",
				AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				Labels:      map[string]string{"env": "prod", "customer": "acme"},
				AppBranchID: explicit.ID,
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
		var install app.Install
		require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
		require.True(s.T(), install.AppBranchID.Valid)
		assert.Equal(s.T(), explicit.ID, install.AppBranchID.String)
	})

	s.Run("explicit app_branch naming another app's branch is rejected", func() {
		otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
		otherBranch := &app.AppBranch{AppID: otherApp.ID, Name: "prod"}
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(otherBranch).Error)

		body := CreateInstallV2Request{
			AppID: s.testApp.ID,
			CreateInstallParams: helpers.CreateInstallParams{
				Name:        "explicit-branch-wrong-app",
				AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
				AppBranchID: otherBranch.ID,
			},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
		assert.Contains(s.T(), rr.Body.String(), "does not belong to this app")
	})
}

func (s *InstallsServiceTestSuite) TestCreateInstallMatchedBranchConfigStates() {
	s.Run("matched branch pins its older locked config when a newer one is still building", func() {
		s.expectQueueCreation()
		branch, lockedCfg := s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)
		s.createBranchConfig(branch, false)

		res := s.createInstallWithLabels("matched-pins-older-locked", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusCreated, res.code, string(res.body))
		install := res.install(s.T())
		assert.Equal(s.T(), lockedCfg.ID, install.AppConfigID)
	})

	s.Run("matched branch still building errors without a fallback", func() {
		branch := &app.AppBranch{AppID: s.testApp.ID, Name: uniqueBranchName("prod")}
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
		_ = s.createBranchConfig(branch, false)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
			AppBranchID: branch.ID,
			InstallGroups: []app.AppBranchInstallGroup{{
				Name:          "prod",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
			}},
		}).Error)

		res := s.createInstallWithLabels("matched-branch-building", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "still building")
		s.assertNoInstallCreated("matched-branch-building")
	})

	s.Run("matched branch with only preview configs errors without a fallback", func() {
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStatePreviewOnly)

		res := s.createInstallWithLabels("matched-branch-preview-only", map[string]string{"env": "prod"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "no active app config")
		s.assertNoInstallCreated("matched-branch-preview-only")
	})
}

func (s *InstallsServiceTestSuite) TestCreateInstallNoMatchAndNoFallback() {
	s.Run("branches exist, no apps-sync config, no labels errors with the no-branch-match wording", func() {
		s.removeFallbackAppConfig()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)

		res := s.createInstallWithLabels("no-sync-config-no-labels", nil)
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "has no app_branch and its labels")
		s.assertNoInstallCreated("no-sync-config-no-labels")
	})

	s.Run("branches exist, no apps-sync config, non-matching labels errors with the no-branch-match wording", func() {
		s.removeFallbackAppConfig()
		s.seedBranch("prod", []app.AppBranchInstallGroup{selectingGroup("prod", labels.Labels{"env": "prod"})}, branchCfgStateActive)

		res := s.createInstallWithLabels("no-sync-config-no-match", map[string]string{"env": "stage"})
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "has no app_branch and its labels")
		s.assertNoInstallCreated("no-sync-config-no-match")
	})

	s.Run("no branches and no apps-sync config keeps the original sync-your-app message", func() {
		s.removeFallbackAppConfig()

		res := s.createInstallWithLabels("no-branches-no-config", nil)
		require.Equal(s.T(), http.StatusBadRequest, res.code, string(res.body))
		assert.Contains(s.T(), string(res.body), "sync your app configuration")
		s.assertNoInstallCreated("no-branches-no-config")
	})
}
