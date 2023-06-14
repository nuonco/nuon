package repos

import (
	"context"
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// create build for repo tests here since Builds follows different paradigm
func createBuild(ctx context.Context, t *testing.T, state repoTestState) *models.Build {
	component := createComponent(ctx, t, state)
	buildID := domains.NewBuildID()
	build := models.Build{
		Model: models.Model{
			ID: buildID,
		},
		GitRef:      "git-ref",
		ComponentID: component.ID,
		CreatedByID: "created-by-id",
	}
	err := state.db.WithContext(ctx).Create(&build).Error
	require.NoError(t, err)
	assert.NotNil(t, build)
	return &build
}

func createDeploy(ctx context.Context, t *testing.T, state repoTestState) *models.Deploy {
	build := createBuild(ctx, t, state)
	install := createInstall(ctx, t, state)

	deployID := shortid.NewNanoID("dpl")
	deploy, err := state.deployRepo.Create(ctx, &models.Deploy{
		BuildID:   build.ID,
		InstallID: install.ID,
		Model:     models.Model{ID: deployID},
	})
	require.NoError(t, err)
	assert.NotNil(t, deploy)
	return deploy
}

func TestGetDeploy(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get a deploy successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeploy := createDeploy(ctx, t, state)

				deploy, err := state.deployRepo.Get(ctx, origDeploy.ID)
				assert.NoError(t, err)
				assert.NotNil(t, deploy)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				deploy := createDeploy(ctx, t, state)

				state.ctxCloseFn()
				fetchedDeploy, err := state.deployRepo.Get(ctx, deploy.ID)
				assert.Error(t, err)
				assert.Nil(t, fetchedDeploy)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				deployID := shortid.NewNanoID("dpl")
				fetchedDeploy, err := state.deployRepo.Get(ctx, deployID)
				assert.Error(t, err)
				assert.Nil(t, fetchedDeploy)
			},
		},
	})
}

func TestUpdateDeploy(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get a deploy successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeploy := createDeploy(ctx, t, state)

				install := createInstall(ctx, t, state)

				origDeploy.InstallID = install.ID

				deploy, err := state.deployRepo.Update(ctx, origDeploy)
				assert.NoError(t, err)
				assert.NotNil(t, deploy)

				dbDeploy, err := state.deployRepo.Get(ctx, origDeploy.ID)
				assert.NoError(t, err)
				assert.NotNil(t, deploy, dbDeploy)
			},
		},
	})
}

func TestListByInstance(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get all deploys successfully when no limit is set",
			fn: func(ctx context.Context, state repoTestState) {
				origDeploy := createDeploy(ctx, t, state)

				deploys, page, err := state.deployRepo.ListByInstance(ctx, origDeploy.InstanceID, &models.ConnectionOptions{})
				assert.Nil(t, err)
				assert.NotEmpty(t, page)
				assert.NotEmpty(t, deploys)

				// NOTE(jm): until we've fixed all bugs cleaning up all database objects from previous
				// runs, we can't guarantee this will be the only app in the list
				// assert.Equal(t, apps[0].ID, origApp.ID)
				// assert.Equals(t, len(apps), 1)
				found := false
				for _, deploy := range deploys {
					if deploy.ID == origDeploy.ID {
						found = true
						break
					}
				}
				assert.True(t, found)
			},
		},
		{
			desc: "should error with a context canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				instanceID := domains.NewInstanceID()
				deploys, page, err := state.deployRepo.ListByInstance(ctx, instanceID, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, deploys)
				assert.Nil(t, page)
			},
		},
	})
}
