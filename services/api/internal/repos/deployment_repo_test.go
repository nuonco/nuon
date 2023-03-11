package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDeployment(ctx context.Context, t *testing.T, state repoTestState) *models.Deployment {
	component := createComponent(ctx, t, state)

	deployment, err := state.deploymentRepo.Create(ctx, &models.Deployment{
		ComponentID: component.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, deployment)
	return deployment
}

func TestGetDeployment(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get a deployment successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeployment := createDeployment(ctx, t, state)

				deployment, err := state.deploymentRepo.Get(ctx, origDeployment.ID)
				assert.NoError(t, err)
				assert.NotNil(t, deployment)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				deployment := createDeployment(ctx, t, state)

				state.ctxCloseFn()
				fetchedDeployment, err := state.deploymentRepo.Get(ctx, deployment.ID)
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedDeployment, err := state.deploymentRepo.Get(ctx, uuid.New())
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
			},
		},
	})
}

func TestListByComponents(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get a list of deployments successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeployment := createDeployment(ctx, t, state)
				origDeployment2 := createDeployment(ctx, t, state)

				uuids := []uuid.UUID{origDeployment.ComponentID, origDeployment2.ComponentID}

				deployments, page, err := state.deploymentRepo.ListByComponents(ctx, uuids, &models.ConnectionOptions{})
				assert.NoError(t, err)
				assert.NotNil(t, deployments)
				assert.NotNil(t, page)
				assert.Equal(t, 2, len(deployments))
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				deployment := createDeployment(ctx, t, state)

				state.ctxCloseFn()
				fetchedDeployment, page, err := state.deploymentRepo.ListByComponents(ctx, []uuid.UUID{deployment.ComponentID}, &models.ConnectionOptions{})
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
				assert.Nil(t, page)
			},
		},
	})
}

func TestListByApps(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get a list of deployments successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeployment := createDeployment(ctx, t, state)
				origDeployment2 := createDeployment(ctx, t, state)

				app := createApp(ctx, t, state)
				origDeployment.Component.AppID = app.ID
				origDeployment2.Component.AppID = app.ID

				uuids := []uuid.UUID{origDeployment.Component.App.ID, origDeployment2.Component.App.ID}

				deployments, page, err := state.deploymentRepo.ListByApps(ctx, uuids, &models.ConnectionOptions{})
				assert.NoError(t, err)
				assert.NotNil(t, deployments)
				assert.NotNil(t, page)
				assert.Equal(t, 2, len(deployments))
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				deployment := createDeployment(ctx, t, state)
				app := createApp(ctx, t, state)
				deployment.Component.AppID = app.ID

				state.ctxCloseFn()
				fetchedDeployment, page, err := state.deploymentRepo.ListByApps(ctx, []uuid.UUID{deployment.Component.AppID}, &models.ConnectionOptions{})
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
				assert.Nil(t, page)
			},
		},
	})
}

func TestListByInstalls(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get a list of deployments successfully",
			fn: func(ctx context.Context, state repoTestState) {
				install := createInstall(ctx, t, state)
				install2 := createInstall(ctx, t, state)
				origDeployment := createDeployment(ctx, t, state)
				origDeployment2 := createDeployment(ctx, t, state)

				// update Component.AppID so it's the same as the Install's
				origDeployment.Component.AppID = install.AppID
				origDeployment2.Component.AppID = install2.AppID
				origDeployment.Component.App = install.App
				origDeployment2.Component.App = install2.App
				_, _ = state.componentRepo.Update(ctx, &origDeployment.Component)
				_, _ = state.componentRepo.Update(ctx, &origDeployment2.Component)

				uuids := []uuid.UUID{install.ID, install2.ID}

				deployments, page, err := state.deploymentRepo.ListByInstalls(ctx, uuids, &models.ConnectionOptions{})
				assert.NoError(t, err)
				assert.NotNil(t, deployments)
				assert.NotNil(t, page)
				assert.Equal(t, 2, len(deployments))
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				deployment := createDeployment(ctx, t, state)
				install := createInstall(ctx, t, state)
				deployment.Component.AppID = install.AppID

				state.ctxCloseFn()
				fetchedDeployment, page, err := state.deploymentRepo.ListByInstalls(ctx, []uuid.UUID{install.ID}, &models.ConnectionOptions{})
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
				assert.Nil(t, page)
			},
		},
	})
}
