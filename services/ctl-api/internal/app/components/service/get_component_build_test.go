package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/generic"
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
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(job).UpdateColumn("status", app.RunnerJobStatusFailed).Error)
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
	require.NotNil(s.T(), response.CompositeError)
	assert.Equal(s.T(), expected.Type, response.CompositeError.Type)
	assert.Equal(s.T(), expected.Severity, response.CompositeError.Severity)
	assert.Equal(s.T(), expected.Message, response.CompositeError.Message)
}

func (s *ComponentsServiceTestSuite) TestGetAppComponentBuildCompositeErrorPrecedence() {
	cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
	ccc := s.getSeededConfigConnection(cmp.ID)

	genericBuildError := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     generic.GenericErrorType,
		Severity: compositeerrors.SeverityError,
		Message:  "build job failed: child workflow execution error",
	}
	specificBuildError := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     "terraform.aws_permission",
		Severity: compositeerrors.SeverityError,
		Message:  "missing permission",
	}
	genericJobError := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     generic.GenericErrorType,
		Severity: compositeerrors.SeverityError,
		Message:  "runner error output",
	}
	specificJobError := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     "terraform.aws_permission",
		Severity: compositeerrors.SeverityError,
		Message:  "runner missing permission",
	}

	cases := map[string]struct {
		buildCompositeError *compositeerrors.CompositeErrorData
		jobCompositeError   *compositeerrors.CompositeErrorData
		expected            *compositeerrors.CompositeErrorData
	}{
		"generic build error is replaced by generic job error": {
			buildCompositeError: genericBuildError,
			jobCompositeError:   genericJobError,
			expected:            genericJobError,
		},
		"non-generic build error is kept over generic job error": {
			buildCompositeError: specificBuildError,
			jobCompositeError:   genericJobError,
			expected:            specificBuildError,
		},
		"generic build error is replaced by non-generic job error": {
			buildCompositeError: genericBuildError,
			jobCompositeError:   specificJobError,
			expected:            specificJobError,
		},
		"build error is kept when the job has no error": {
			buildCompositeError: specificBuildError,
			jobCompositeError:   nil,
			expected:            specificBuildError,
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			seededBuild := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), ccc.ID)
			if tc.buildCompositeError != nil {
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(seededBuild).UpdateColumn("composite_error", tc.buildCompositeError).Error)
			}
			if tc.jobCompositeError != nil {
				job := s.deps.Seeder.CreateRunnerJob(s.ctx, s.T(), seededBuild.ID, "component_builds")
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(job).UpdateColumn("status", app.RunnerJobStatusFailed).Error)
				execution := &app.RunnerJobExecution{
					RunnerJobID: job.ID,
					Status:      app.RunnerJobExecutionStatusFailed,
				}
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(execution).Error)
				result := &app.RunnerJobExecutionResult{
					RunnerJobExecutionID: execution.ID,
					CompositeError:       tc.jobCompositeError,
				}
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(result).Error)
			}

			path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", s.testApp.ID, cmp.ID, seededBuild.ID)
			rr := s.makeRequest(http.MethodGet, path, nil)
			require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

			var response app.ComponentBuild
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
			expected := tc.expected
			if expected != nil {
				raw, err := json.Marshal(expected)
				require.NoError(s.T(), err)
				expected = &compositeerrors.CompositeErrorData{}
				require.NoError(s.T(), json.Unmarshal(raw, expected))
			}
			require.NotNil(s.T(), response.CompositeError)
			assert.Equal(s.T(), expected, response.CompositeError)
		})
	}
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
