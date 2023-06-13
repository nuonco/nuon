package repos

import (
	"context"
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
)

// createInstall
func createInstance(ctx context.Context, t *testing.T, state repoTestState) []*models.Instance {
	deploy := createDeploy(ctx, t, state)

	instance, err := state.instanceRepo.Create(ctx, []*models.Instance{&models.Instance{
		ComponentID: deploy.Build.ComponentID,
		InstallID:   deploy.InstallID,
		BuildID:     deploy.BuildID,
	}})
	assert.Nil(t, err)
	assert.NotNil(t, instance)
	return instance
}

func TestGetInstance(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get an instance successfully",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				instance := instances[0]

				instance, err := state.instanceRepo.Get(ctx, instance.ID)
				assert.Nil(t, err)
				assert.NotNil(t, instance)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				instance := instances[0]

				state.ctxCloseFn()
				fetchedInstall, err := state.instanceRepo.Get(ctx, instance.ID)
				assert.Nil(t, fetchedInstall)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedInstance, err := state.instanceRepo.Get(ctx, domains.NewInstanceID())
				assert.Nil(t, fetchedInstance)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetByInstallAndComponentInstance(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get an instance successfully",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				instance := instances[0]

				instance, err := state.instanceRepo.GetByInstallAndComponent(ctx, instance.InstallID, instance.ComponentID)
				assert.Nil(t, err)
				assert.NotNil(t, instance)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				instance := instances[0]

				state.ctxCloseFn()
				instance, err := state.instanceRepo.GetByInstallAndComponent(ctx, instance.InstallID, instance.ComponentID)
				assert.Nil(t, instance)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				instance, err := state.instanceRepo.Get(ctx, domains.NewInstanceID())
				assert.Nil(t, instance)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestCreateInstance(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should upsert an instance successfully",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				instance := instances[0]

				deploy := createDeploy(ctx, t, state)

				instance.BuildID = deploy.BuildID

				instancesUpdate, err := state.instanceRepo.Create(ctx, []*models.Instance{instance})
				instanceUpdate := instancesUpdate[0]
				assert.Nil(t, err)
				assert.NotNil(t, instanceUpdate)
				assert.Equal(t, instance.ID, instanceUpdate.ID)
				assert.Equal(t, instance.BuildID, instanceUpdate.BuildID)
			},
		},
	})
}

func TestListByInstall(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should list instances successfully",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				secondInstance := createInstance(ctx, t, state)
				secondInstance[0].InstallID = instances[0].InstallID
				instances = append(instances, secondInstance[0])

				res, err := state.instanceRepo.ListByInstall(ctx, instances[0].InstallID)
				assert.Nil(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, len(instances), len(res))
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				instances := createInstance(ctx, t, state)
				instance := instances[0]

				state.ctxCloseFn()
				fetchedInstall, err := state.instanceRepo.ListByInstall(ctx, instance.InstallID)
				assert.Nil(t, fetchedInstall)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedInstance, err := state.instanceRepo.ListByInstall(ctx, domains.NewInstallID())
				assert.Nil(t, fetchedInstance)
				assert.NotNil(t, err)
			},
		},
	})
}
