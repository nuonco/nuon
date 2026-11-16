package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/orgfeatures"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestCreateInstallV2Success() {
	s.expectQueueCreation()

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:       "my-install",
			AWSAccount: &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))

	assert.NotEmpty(s.T(), install.ID)
	assert.Equal(s.T(), "my-install", install.Name)
	assert.Equal(s.T(), s.testApp.ID, install.AppID)

	captured := tests.GetQueueSignals(s.T(), s.deps.DB)
	require.GreaterOrEqual(s.T(), len(captured), 4)

	var signalTypes []string
	for _, c := range captured {
		signalTypes = append(signalTypes, string(c.Type))
	}
	assert.Contains(s.T(), signalTypes, "install-created")
	assert.Contains(s.T(), signalTypes, "poll-dependencies")
	assert.Contains(s.T(), signalTypes, "execute-workflow")
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithInputs() {
	s.expectQueueCreation()

	region := "us-east-1"
	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:       "install-with-inputs",
			AWSAccount: &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
			Inputs: map[string]*string{
				"region": &region,
			},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "install-with-inputs", install.Name)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2AzureAccount() {
	s.expectQueueCreation()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppRunnerConfig{}).
		Where(app.AppRunnerConfig{AppID: s.testApp.ID}).
		Update("type", app.AppRunnerTypeAzure).Error)

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:         "azure-install",
			AzureAccount: &helpers.CreateInstallAzureAccountParams{Location: "eastus"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "azure-install", install.Name)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2ValidationErrors() {
	testCases := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing app_id",
			body: map[string]interface{}{
				"name":        "foo",
				"aws_account": map[string]string{"region": "us-west-2"},
			},
		},
		{
			name: "missing name",
			body: map[string]interface{}{
				"app_id":      s.testApp.ID,
				"aws_account": map[string]string{"region": "us-west-2"},
			},
		},
		{
			name: "no cloud account",
			body: map[string]interface{}{
				"app_id": s.testApp.ID,
				"name":   "foo",
			},
		},
		{
			name: "aws missing region",
			body: map[string]interface{}{
				"app_id":      s.testApp.ID,
				"name":        "foo",
				"aws_account": map[string]string{},
			},
		},
		{
			name: "empty body",
			body: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodPost, "/v1/installs", tc.body)
			require.Equal(s.T(), http.StatusBadRequest, rr.Code,
				"expected 400 for %s, got %d: %s", tc.name, rr.Code, rr.Body.String())
		})
	}
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2InvalidAppID() {
	body := CreateInstallV2Request{
		AppID: "app_nonexistent_000000000",
		CreateInstallParams: helpers.CreateInstallParams{
			Name:       "bad-app",
			AWSAccount: &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateInstallDeprecatedRoute() {
	s.expectQueueCreation()

	body := helpers.CreateInstallParams{
		Name:       "deprecated-route-install",
		AWSAccount: &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
	}

	path := fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "deprecated-route-install", install.Name)
	assert.Equal(s.T(), s.testApp.ID, install.AppID)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2AllowsUnbranchedInstallWhenAppHasBranches() {
	s.expectQueueCreation()

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:       "branched-missing-ids",
			AWSAccount: &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.False(s.T(), install.AppBranchID.Valid)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2PersistsExplicitBranchOwnership() {
	s.expectQueueCreation()

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	branchConfig := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(branchConfig).
		Update("app_branch_id", branch.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
		AppBranchID:   branch.ID,
		InstallGroups: []app.AppBranchInstallGroup{{Name: "Default", Order: 0, Default: true}},
	}).Error)

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:        "branch-owned",
			AppBranchID: branch.ID,
			AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	require.True(s.T(), install.AppBranchID.Valid)
	assert.Equal(s.T(), branch.ID, install.AppBranchID.String)

	var connection app.InstallAppBranchConnection
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallAppBranchConnection{
			InstallID:   install.ID,
			AppBranchID: branch.ID,
			Active:      true,
		}).
		First(&connection).Error)
}

func (s *InstallsServiceTestSuite) createBranchConfig(branch *app.AppBranch, locked bool) *app.AppConfig {
	cfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(cfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_pinned": true, "builds_locked": locked}).Error)
	return cfg
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2RejectsBranchStillBuilding() {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	s.createBranchConfig(branch, false)

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:        "branch-building",
			AppBranchID: branch.ID,
			AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Contains(s.T(), rr.Body.String(), "still building")
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2PinsLatestLockedBranchConfig() {
	s.expectQueueCreation()

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	locked := s.createBranchConfig(branch, true)
	s.createBranchConfig(branch, false)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
		AppBranchID:   branch.ID,
		InstallGroups: []app.AppBranchInstallGroup{{Name: "Default", Order: 0, Default: true}},
	}).Error)

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:        "branch-locked",
			AppBranchID: branch.ID,
			AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), locked.ID, install.AppConfigID)
}

func (s *InstallsServiceTestSuite) createInstallWithInputs(appBranchID string, inputs map[string]*string) *httptest.ResponseRecorder {
	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:        "acme",
			AppBranchID: appBranchID,
			AWSAccount:  &helpers.CreateInstallAWSAccountParams{Region: "us-west-2"},
			Inputs:      inputs,
		},
	}
	return s.makeRequest(http.MethodPost, "/v1/installs", body)
}

func (s *InstallsServiceTestSuite) errDescription(rr *httptest.ResponseRecorder) string {
	var resp stderr.ErrResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp), rr.Body.String())
	return resp.Description
}

const (
	unknownInputSyncOrBranchRun = `Unknown input "retired": the app config doesn't define it. Add it, then run ` + "`nuon apps sync`" + ` or run your app branch.`
	unknownInputBranchRun       = `Unknown input "retired": the app config doesn't define it. Add it, then run your app branch.`
)

func (s *InstallsServiceTestSuite) TestCreateInstallV2UnknownInputWithoutBranchSuggestsAppSyncOrBranchRun() {
	val := "1"
	rr := s.createInstallWithInputs("", map[string]*string{"retired": &val})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), unknownInputSyncOrBranchRun, s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2UnknownInputOnBranchSuggestsAppSyncOrBranchRun() {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	s.createBranchConfig(branch, true)

	val := "1"
	rr := s.createInstallWithInputs(branch.ID, map[string]*string{"retired": &val})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), unknownInputSyncOrBranchRun, s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) addRequiredInput(appConfigID string) {
	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.
		Preload("AppInputs").
		Where(app.AppInputConfig{AppConfigID: appConfigID}).
		First(&inputCfg).Error)
	require.NotEmpty(s.T(), inputCfg.AppInputs)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppInput{
		AppInputConfigID: inputCfg.ID,
		AppInputGroupID:  inputCfg.AppInputs[0].AppInputGroupID,
		Name:             "size",
		DisplayName:      "Size",
		Type:             app.AppInputTypeString,
		Required:         true,
		Source:           app.AppInputSourceVendor,
	}).Error)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2MissingRequiredInputOmitsUnknownInputHint() {
	s.addRequiredInput(s.testAppConfig.ID)

	rr := s.createInstallWithInputs("", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "size is required, please add a value value for the input", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2MissingRequiredInputWhenAppSyncDisabledOmitsUnknownInputHint() {
	s.setOrgFeatures(app.OrgFeatureDisableAppSync)
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	branchConfig := s.createBranchConfig(branch, true)
	s.addRequiredInput(branchConfig.ID)

	rr := s.createInstallWithInputs(branch.ID, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "size is required, please add a value value for the input", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2UnknownInputOnBranchWhenAppSyncDisabledSuggestsBranchRun() {
	s.setOrgFeatures(app.OrgFeatureDisableAppSync)
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	s.createBranchConfig(branch, true)

	val := "1"
	rr := s.createInstallWithInputs(branch.ID, map[string]*string{"retired": &val})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), unknownInputBranchRun, s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithoutBranchAndNoActiveConfigAsksForSyncOrBranch() {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(s.testAppConfig).
		Update("app_branch_id", branch.ID).Error)

	rr := s.createInstallWithInputs("", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "No sync config exists for this app. Run `nuon apps sync`, or choose an app branch.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithoutBranchWhenAppSyncDisabledRequiresAppBranch() {
	s.setOrgFeatures(app.OrgFeatureDisableAppSync)

	rr := s.createInstallWithInputs("", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "App sync is disabled for this org. Choose an app branch for this install.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithoutBranchWhenAppSyncDisabledByDeploymentRequiresAppBranch() {
	orgfeatures.SetForced(string(app.OrgFeatureDisableAppSync))
	s.T().Cleanup(func() { orgfeatures.SetForced("") })

	rr := s.createInstallWithInputs("", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "App sync is disabled for this org. Choose an app branch for this install.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithBranchWhenAppSyncDisabled() {
	s.expectQueueCreation()
	s.setOrgFeatures(app.OrgFeatureDisableAppSync)

	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	branchConfig := s.createBranchConfig(branch, true)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
		AppBranchID:   branch.ID,
		InstallGroups: []app.AppBranchInstallGroup{{Name: "Default", Order: 0, Default: true}},
	}).Error)

	rr := s.createInstallWithInputs(branch.ID, nil)
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), branchConfig.ID, install.AppConfigID)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithoutAnyConfigAsksForSync() {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(s.testAppConfig).Error)

	rr := s.createInstallWithInputs("", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "No sync config exists for this app. Run `nuon apps sync`, or choose an app branch.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2ExplicitBranchWithoutActiveConfig() {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	rr := s.createInstallWithInputs(branch.ID, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "The selected app branch has no app config yet. Run the app branch first.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2ExplicitBranchWithoutActiveConfigWhenAppSyncDisabled() {
	s.setOrgFeatures(app.OrgFeatureDisableAppSync)
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	rr := s.createInstallWithInputs(branch.ID, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "The selected app branch has no app config yet. Run the app branch first.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2NoSandboxOrRunnerConfig() {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(&app.AppSandboxConfig{}, "app_id = ?", s.testApp.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(&app.AppRunnerConfig{}, "app_id = ?", s.testApp.ID).Error)

	rr := s.createInstallWithInputs("", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "No sandbox or runner config found. Add them, then run `nuon apps sync` or run your app branch.", s.errDescription(rr))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2NoSandboxOrRunnerConfigWhenAppSyncDisabled() {
	s.setOrgFeatures(app.OrgFeatureDisableAppSync)
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: "prod"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)
	s.createBranchConfig(branch, true)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(&app.AppSandboxConfig{}, "app_id = ?", s.testApp.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(&app.AppRunnerConfig{}, "app_id = ?", s.testApp.ID).Error)

	rr := s.createInstallWithInputs(branch.ID, nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Equal(s.T(), "No sandbox or runner config found. Add them, then run your app branch.", s.errDescription(rr))
}
