package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestCreateComponentDeploySuccess() {
	install := s.createTestInstallWithActiveRunner()

	component := s.getSeededComponent(app.ComponentTypeHelmChart)
	var ccc app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.Where(app.ComponentConfigConnection{
		ComponentID: component.ID, AppConfigID: s.testAppConfig.ID,
	}).First(&ccc).Error)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, ccc.ComponentID)

	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
	require.NoError(s.T(), s.deps.DB.Model(build).Update("status", app.ComponentBuildStatusActive).Error)
	newerBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
	require.NoError(s.T(), s.deps.DB.Model(newerBuild).Update("status", app.ComponentBuildStatusActive).Error)
	require.NoError(s.T(), s.deps.DB.Model(&ccc).Update("latest_build_id", newerBuild.ID).Error)

	dependency := s.getSeededComponent(app.ComponentTypeExternalImage)
	var dependencyConfig app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.Where(app.ComponentConfigConnection{
		ComponentID: dependency.ID, AppConfigID: s.testAppConfig.ID,
	}).First(&dependencyConfig).Error)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, dependency.ID)
	dependencyBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), dependencyConfig.ID)
	require.NoError(s.T(), s.deps.DB.Model(dependencyBuild).Update("status", app.ComponentBuildStatusActive).Error)
	require.NoError(s.T(), s.deps.DB.Model(&dependencyConfig).Update("latest_build_id", dependencyBuild.ID).Error)
	ccc.ComponentDependencyIDs = []string{dependency.ID}
	require.NoError(s.T(), s.deps.DB.Model(&ccc).Select("component_dependency_ids").Updates(&ccc).Error)

	body := CreateInstallDeployRequest{
		BuildID: build.ID, DeployDependencies: true,
	}

	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, ccc.ComponentID)
	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var deploy app.InstallDeploy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &deploy))
	assert.NotEmpty(s.T(), deploy.ID)
	assert.Equal(s.T(), build.ID, deploy.ComponentBuildID)

	// Workflow should have been created.
	require.NotNil(s.T(), deploy.WorkflowID)
	workflowID := *deploy.WorkflowID
	assert.NotEmpty(s.T(), workflowID)

	// Verify deploy persisted in DB.
	var dbDeploy app.InstallDeploy
	require.NoError(s.T(), s.deps.DB.Where("id = ?", deploy.ID).First(&dbDeploy).Error)
	assert.Equal(s.T(), build.ID, dbDeploy.ComponentBuildID)

	// Verify workflow persisted in DB.
	var dbWorkflow app.Workflow
	require.NoError(s.T(), s.deps.DB.Where("id = ?", workflowID).First(&dbWorkflow).Error)
	assert.Equal(s.T(), install.ID, dbWorkflow.OwnerID)
	require.Len(s.T(), tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, workflowID), 1)

	worker := activities.New(activities.Params{
		DB: s.deps.DB, V: s.deps.V, Helpers: s.installsService.helpers,
		AppsHelpers: s.installsService.appsHelpers, RunnersHelpers: s.installsService.runnersHelpers,
		Audit: s.installsService.audit,
	})
	retryRequest := activities.CreateInstallDeployRequest{
		InstallID: install.ID, ComponentID: component.ID, BuildID: build.ID,
		Type: app.InstallDeployTypeApply, WorkflowID: workflowID,
	}
	retry, err := worker.CreateInstallDeploy(s.ctx, retryRequest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), build.ID, retry.ComponentBuildID)
	wrongPrimary := retryRequest
	wrongPrimary.BuildID = newerBuild.ID
	_, err = worker.CreateInstallDeploy(s.ctx, wrongPrimary)
	require.Error(s.T(), err)

	childRequest := activities.CreateInstallDeployRequest{
		InstallID: install.ID, ComponentID: dependency.ID, BuildID: dependencyBuild.ID,
		Type: app.InstallDeployTypeSync, WorkflowID: workflowID,
	}
	child, err := worker.CreateInstallDeploy(s.ctx, childRequest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), dependencyBuild.ID, child.ComponentBuildID)
	require.Equal(s.T(), workflowID, *child.InstallWorkflowID)

	newerDependencyBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), dependencyConfig.ID)
	require.NoError(s.T(), s.deps.DB.Model(newerDependencyBuild).Update("status", app.ComponentBuildStatusActive).Error)
	require.NoError(s.T(), s.deps.DB.Model(&dependencyConfig).Update("latest_build_id", newerDependencyBuild.ID).Error)
	_, err = worker.CreateInstallDeploy(s.ctx, childRequest)
	require.Error(s.T(), err)

	var runnerGroup app.RunnerGroup
	require.NoError(s.T(), s.deps.DB.Where(app.RunnerGroup{OwnerID: install.ID, OwnerType: "installs"}).First(&runnerGroup).Error)
	var runner app.Runner
	require.NoError(s.T(), s.deps.DB.Where(app.Runner{RunnerGroupID: runnerGroup.ID}).First(&runner).Error)
	logStream := app.LogStream{OwnerID: deploy.ID, OwnerType: "install_deploys"}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&logStream).Error)
	jobRequest := activities.CreateDeployJobRequest{
		RunnerID: runner.ID, DeployID: retry.ID, LogStreamID: logStream.ID,
		Type: app.RunnerJobTypeHelmChartDeploy, Op: app.RunnerJobOperationTypeApplyPlan,
		Metadata: map[string]string{},
	}
	job, err := worker.CreateDeployJob(s.ctx, &jobRequest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), retry.ID, job.OwnerID)
	require.Equal(s.T(), app.RunnerJobStatusQueued, job.Status)
	syncRequest := activities.CreateSyncJobRequest{
		RunnerID: runner.ID, DeployID: child.ID, LogStreamID: logStream.ID,
		Type: app.RunnerJobTypeOCISync, Op: app.RunnerJobOperationTypeExec,
		Metadata: map[string]string{},
	}
	syncJob, err := worker.CreateSyncJob(s.ctx, &syncRequest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), child.ID, syncJob.OwnerID)
	require.Equal(s.T(), app.RunnerJobStatusQueued, syncJob.Status)
	childRetry := childRequest
	childRetry.BuildID = newerDependencyBuild.ID
	_, err = worker.CreateInstallDeploy(s.ctx, childRetry)
	require.NoError(s.T(), err)

	newConfig := s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.Model(&app.Install{}).Where(app.Install{ID: install.ID}).Update("app_config_id", newConfig.ID).Error)
	for _, request := range []activities.CreateInstallDeployRequest{retryRequest, childRetry} {
		_, err = worker.CreateInstallDeploy(s.ctx, request)
		require.Error(s.T(), err)
	}
	_, err = worker.CreateDeployJob(s.ctx, &jobRequest)
	require.Error(s.T(), err)
	_, err = worker.CreateSyncJob(s.ctx, &syncRequest)
	require.Error(s.T(), err)
	for _, ownerID := range []string{retry.ID, child.ID} {
		var jobs int64
		require.NoError(s.T(), s.deps.DB.Model(&app.RunnerJob{}).Where(app.RunnerJob{OwnerID: ownerID}).Count(&jobs).Error)
		require.EqualValues(s.T(), 1, jobs)
	}
	var deploys int64
	require.NoError(s.T(), s.deps.DB.Model(&app.InstallDeploy{}).Where(app.InstallDeploy{InstallWorkflowID: &workflowID}).Count(&deploys).Error)
	require.EqualValues(s.T(), 4, deploys)
}

func (s *InstallsServiceTestSuite) TestCreateComponentDeployRequestReplay() {
	install := s.createTestInstallWithActiveRunner()
	component := s.getSeededComponent(app.ComponentTypeHelmChart)
	var ccc app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.Where(app.ComponentConfigConnection{
		ComponentID: component.ID, AppConfigID: s.testAppConfig.ID,
	}).First(&ccc).Error)
	installed := s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, component.ID)
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
	require.NoError(s.T(), s.deps.DB.Model(build).Update("status", app.ComponentBuildStatusActive).Error)
	body := CreateInstallDeployRequest{RequestID: "deploy-" + install.ID, BuildID: build.ID}
	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, component.ID)

	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())
	var original app.InstallDeploy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &original))
	require.NotEmpty(s.T(), original.ID)
	require.NotNil(s.T(), original.WorkflowID)
	require.NotEmpty(s.T(), *original.WorkflowID)

	rr = s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())
	var replay app.InstallDeploy
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &replay))
	require.Equal(s.T(), original.ID, replay.ID)
	require.Equal(s.T(), original.WorkflowID, replay.WorkflowID)
	require.Equal(s.T(), build.ID, replay.ComponentBuildID)

	changed := body
	changed.PlanOnly = true
	rr = s.makeRequest(http.MethodPost, path, changed)
	require.Equal(s.T(), http.StatusConflict, rr.Code, "body: %s", rr.Body.String())

	newConfig := s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.Model(&app.Install{}).Where(app.Install{ID: install.ID}).Update("app_config_id", newConfig.ID).Error)
	rr = s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusConflict, rr.Code, "body: %s", rr.Body.String())

	var deploys []app.InstallDeploy
	require.NoError(s.T(), s.deps.DB.Where(app.InstallDeploy{InstallComponentID: installed.ID}).Find(&deploys).Error)
	require.Len(s.T(), deploys, 1)
	require.Equal(s.T(), original.ID, deploys[0].ID)
	require.Equal(s.T(), build.ID, deploys[0].ComponentBuildID)
	require.Equal(s.T(), s.testAppConfig.ID, deploys[0].AppConfigID)
	require.Equal(s.T(), original.WorkflowID, deploys[0].InstallWorkflowID)

	var workflows []app.Workflow
	require.NoError(s.T(), s.deps.DB.Where(app.Workflow{OwnerID: install.ID, Type: app.WorkflowTypeManualDeploy}).Find(&workflows).Error)
	require.Len(s.T(), workflows, 1)
	require.Equal(s.T(), *original.WorkflowID, workflows[0].ID)
	require.False(s.T(), workflows[0].PlanOnly)
	require.NotNil(s.T(), workflows[0].Request)
	require.Equal(s.T(), body.RequestID, workflows[0].Request.RequestID)
	require.Equal(s.T(), s.testAppConfig.ID, workflows[0].Request.PinnedAppConfigID)
	signals := tests.GetQueueSignalsByOwner(s.T(), s.deps.DB, workflows[0].ID)
	require.Len(s.T(), signals, 1)
	require.Equal(s.T(), "execute-workflow", string(signals[0].Type))
}

func (s *InstallsServiceTestSuite) TestCreateComponentDeployNotFound() {
	ccc := s.testAppConfig.ComponentConfigConnections[0]

	body := CreateInstallDeployRequest{
		BuildID: "fake-build",
	}

	path := fmt.Sprintf("/v1/installs/nonexistent/components/%s/deploys", ccc.ComponentID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateComponentDeployInvalidBuild() {
	install := s.createTestInstall()
	ccc := s.testAppConfig.ComponentConfigConnections[0]

	body := CreateInstallDeployRequest{
		BuildID: "nonexistent-build",
	}

	path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, ccc.ComponentID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateComponentDeployRejectsUnrelatedBuild() {
	install := s.createTestInstallWithActiveRunner()
	component := s.getSeededComponent(app.ComponentTypeHelmChart)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, component.ID)

	otherComponent := s.getSeededComponent(app.ComponentTypeTerraformModule)
	var otherConfig app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.Where(app.ComponentConfigConnection{
		ComponentID: otherComponent.ID, AppConfigID: s.testAppConfig.ID,
	}).First(&otherConfig).Error)
	otherBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), otherConfig.ID)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, otherComponent.ID)

	otherApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	otherAppConfig := s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), otherApp.ID)
	foreignComponent := s.deps.Seeder.CreateComponent(s.ctx, s.T(), otherApp.ID, app.ComponentTypeHelmChart)
	foreignConfig := s.deps.Seeder.CreateHelmComponentConfigConnection(s.ctx, s.T(), foreignComponent.ID, otherAppConfig.ID)
	foreignBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), foreignConfig.ID)
	s.deps.Seeder.CreateInstallComponent(s.ctx, s.T(), install.ID, foreignComponent.ID)

	for name, tc := range map[string]struct {
		componentID string
		buildID     string
	}{
		"wrong component": {component.ID, otherBuild.ID},
		"wrong app":       {foreignComponent.ID, foreignBuild.ID},
	} {
		s.Run(name, func() {
			path := fmt.Sprintf("/v1/installs/%s/components/%s/deploys", install.ID, tc.componentID)
			rr := s.makeRequest(http.MethodPost, path, CreateInstallDeployRequest{BuildID: tc.buildID})
			require.Equal(s.T(), http.StatusBadRequest, rr.Code, "body: %s", rr.Body.String())
			var deploys, workflows int64
			require.NoError(s.T(), s.deps.DB.Model(&app.InstallDeploy{}).Where(app.InstallDeploy{OrgID: s.testOrg.ID}).Count(&deploys).Error)
			require.NoError(s.T(), s.deps.DB.Model(&app.Workflow{}).Where(app.Workflow{OwnerID: install.ID}).Count(&workflows).Error)
			require.Zero(s.T(), deploys)
			require.Zero(s.T(), workflows)
			require.Empty(s.T(), tests.GetQueueSignals(s.T(), s.deps.DB))
		})
	}
}
