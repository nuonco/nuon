package repos

import (
	"context"
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDeployment(ctx context.Context, t *testing.T, state repoTestState) *models.Deployment {
	component := createComponent(ctx, t, state)
	deploymentID := domains.NewDeploymentID()

	deployment, err := state.deploymentRepo.Create(ctx, &models.Deployment{
		ComponentID: component.ID,
		Model:       models.Model{ID: deploymentID},
	})
	require.NoError(t, err)
	assert.NotNil(t, deployment)
	return deployment
}

func TestGetDeployment(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

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
				deploymentID := domains.NewDeploymentID()
				fetchedDeployment, err := state.deploymentRepo.Get(ctx, deploymentID)
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
			},
		},
	})
}

func TestListByComponents(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get a list of deployments successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeployment := createDeployment(ctx, t, state)
				origDeployment2 := createDeployment(ctx, t, state)

				componentIDs := []string{origDeployment.ComponentID, origDeployment2.ComponentID}

				deployments, page, err := state.deploymentRepo.ListByComponents(ctx, componentIDs, &models.ConnectionOptions{})
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
				fetchedDeployment, page, err := state.deploymentRepo.ListByComponents(ctx, []string{deployment.ComponentID}, &models.ConnectionOptions{})
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
				assert.Nil(t, page)
			},
		},
	})
}

func TestListByApps(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get a list of deployments successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origDeployment := createDeployment(ctx, t, state)
				origDeployment2 := createDeployment(ctx, t, state)

				app := createApp(ctx, t, state)
				origDeployment.Component.AppID = app.ID
				origDeployment2.Component.AppID = app.ID

				appIDs := []string{origDeployment.Component.App.ID, origDeployment2.Component.App.ID}

				deployments, page, err := state.deploymentRepo.ListByApps(ctx, appIDs, &models.ConnectionOptions{})
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
				fetchedDeployment, page, err := state.deploymentRepo.ListByApps(ctx, []string{deployment.Component.AppID}, &models.ConnectionOptions{})
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
				assert.Nil(t, page)
			},
		},
	})
}

func TestListByInstalls(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

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

				installIDs := []string{install.ID, install2.ID}

				deployments, page, err := state.deploymentRepo.ListByInstalls(ctx, installIDs, &models.ConnectionOptions{})
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
				fetchedDeployment, page, err := state.deploymentRepo.ListByInstalls(ctx, []string{install.ID}, &models.ConnectionOptions{})
				assert.Error(t, err)
				assert.Nil(t, fetchedDeployment)
				assert.Nil(t, page)
			},
		},
	})
}
