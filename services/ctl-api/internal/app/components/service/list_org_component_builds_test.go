package service

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *ComponentsServiceTestSuite) TestListOrgComponentBuilds() {
	component := s.getSeededComponent(app.ComponentTypeHelmChart)
	config := s.getSeededConfigConnection(component.ID)
	oldest := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), config.ID)
	jobBacked := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), config.ID)
	newest := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), config.ID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	s.setComponentBuildCreatedAt(oldest, now.Add(-2*time.Minute))
	s.setComponentBuildCreatedAt(jobBacked, now.Add(-time.Minute))
	s.setComponentBuildCreatedAt(newest, now)

	s.createComponentBuildJob(jobBacked.ID, app.RunnerJobGroupBuild, app.RunnerJobOperationTypeBuild, app.RunnerJobTypeHelmChartBuild)
	executionJob := s.createComponentBuildJob(jobBacked.ID, app.RunnerJobGroupBuild, app.RunnerJobOperationTypeBuild, app.RunnerJobTypeHelmChartBuild)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(executionJob).UpdateColumn("created_at", now.Add(time.Minute)).Error)
	s.createComponentBuildJob(jobBacked.ID, app.RunnerJobGroupSync, app.RunnerJobOperationTypeExec, app.RunnerJobTypeFetchImageMetadata)

	rr := s.makeRequest(http.MethodGet, "/v1/component-builds?limit=2", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var firstPage OrgComponentBuildHistoryResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &firstPage))
	require.Len(s.T(), firstPage.Items, 2)
	require.NotNil(s.T(), firstPage.NextCursor)
	assert.Nil(s.T(), firstPage.PreviousCursor)
	assert.Equal(s.T(), newest.ID, firstPage.Items[0].Build.ID)
	assert.Nil(s.T(), firstPage.Items[0].BuildRunnerJobID)
	assert.Equal(s.T(), jobBacked.ID, firstPage.Items[1].Build.ID)
	require.NotNil(s.T(), firstPage.Items[1].BuildRunnerJobID)
	assert.Equal(s.T(), executionJob.ID, *firstPage.Items[1].BuildRunnerJobID)
	assert.Equal(s.T(), s.testApp.ID, firstPage.Items[0].AppID)
	assert.Equal(s.T(), component.ID, firstPage.Items[0].ComponentID)

	inserted := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), config.ID)
	s.setComponentBuildCreatedAt(inserted, now.Add(time.Minute))

	rr = s.makeRequest(http.MethodGet, "/v1/component-builds?limit=2&cursor="+url.QueryEscape(*firstPage.NextCursor), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var secondPage OrgComponentBuildHistoryResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &secondPage))
	require.Len(s.T(), secondPage.Items, 1)
	assert.Equal(s.T(), oldest.ID, secondPage.Items[0].Build.ID)
	assert.NotEqual(s.T(), inserted.ID, secondPage.Items[0].Build.ID)
	require.NotNil(s.T(), secondPage.PreviousCursor)
	assert.Nil(s.T(), secondPage.NextCursor)

	rr = s.makeRequest(http.MethodGet, "/v1/component-builds?limit=2&cursor="+url.QueryEscape(*secondPage.PreviousCursor), nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var previousPage OrgComponentBuildHistoryResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &previousPage))
	require.Len(s.T(), previousPage.Items, 2)
	assert.Equal(s.T(), newest.ID, previousPage.Items[0].Build.ID)
	assert.Equal(s.T(), jobBacked.ID, previousPage.Items[1].Build.ID)
}

func (s *ComponentsServiceTestSuite) TestListOrgComponentBuildsRejectsInvalidCursor() {
	rr := s.makeRequest(http.MethodGet, "/v1/component-builds?cursor=invalid", nil)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
}

func (s *ComponentsServiceTestSuite) createComponentBuildJob(buildID string, group app.RunnerJobGroup, operation app.RunnerJobOperationType, typ app.RunnerJobType) *app.RunnerJob {
	job := s.deps.Seeder.CreateRunnerJob(s.ctx, s.T(), buildID, "component_builds")
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(job).Updates(map[string]any{
		"status":             app.RunnerJobStatusAvailable,
		"status_description": "test job",
		"group":              group,
		"operation":          operation,
		"type":               typ,
		"executor":           app.RunnerJobExecutorControlPlane,
		"queue_timeout":      5 * time.Minute,
		"available_timeout":  10 * time.Minute,
		"execution_timeout":  30 * time.Minute,
		"overall_timeout":    45 * time.Minute,
	}).Error)
	return job
}

func (s *ComponentsServiceTestSuite) setComponentBuildCreatedAt(build *app.ComponentBuild, createdAt time.Time) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(build).UpdateColumn("created_at", createdAt).Error)
}
