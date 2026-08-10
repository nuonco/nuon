package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildSuccess() {
	s.Run("returns build by id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
		ccc := s.getSeededConfigConnection(cmp.ID)
		seededBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", s.testApp.ID, cmp.ID, seededBuild.ID)
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

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildIncludesLatestCompositeError() {
	cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
	ccc := s.getSeededConfigConnection(cmp.ID)
	seededBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
	job := s.deps.Seeder.CreateRunnerJob(s.ctx, s.T(), seededBuild.ID, "component_builds")
	execution := &app.RunnerJobExecution{
		RunnerJobID: job.ID,
		Status:      app.RunnerJobExecutionStatusFailed,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(execution).Error)

	expected := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     "terraform.aws_permission",
		Severity: compositeerrors.SeverityError,
		Message:  "missing permission",
	}
	result := &app.RunnerJobExecutionResult{
		RunnerJobExecutionID: execution.ID,
		CompositeError:       expected,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(result).Error)

	path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", s.testApp.ID, cmp.ID, seededBuild.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var response app.ComponentBuild
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
	assert.Equal(s.T(), expected, response.CompositeError)
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildNotFound() {
	s.Run("nonexistent build id", func() {
		cmp := s.getSeededComponent(app.ComponentTypeHelmChart)

		path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", s.testApp.ID, cmp.ID, "bld_nonexistent00000000000")
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusNotFound {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
