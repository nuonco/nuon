package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *ComponentsServiceTestSuite) TestCreateAppDockerBuildConfigRejected() {
	comp := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeDockerBuild)

	path := fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", s.testApp.ID, comp.ID)
	rr := s.makeRequest(http.MethodPost, path, CreateDockerBuildComponentConfigRequest{
		AppConfigID: s.testAppConfig.ID,
		Dockerfile:  "Dockerfile",
		basicVCSConfigRequest: basicVCSConfigRequest{
			PublicGitVCSConfig: &PublicGitVCSConfigRequest{
				Repo:      "owner/repo",
				Directory: ".",
				Branch:    "main",
			},
		},
	})

	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "docker_build components have been deprecated")

	var count int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.ComponentConfigConnection{}).
		Where(&app.ComponentConfigConnection{ComponentID: comp.ID}).
		Count(&count).Error)
	assert.Zero(s.T(), count)
	assert.Empty(s.T(), tests.GetQueueSignals(s.T(), s.deps.DB))
}
