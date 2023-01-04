package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
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
