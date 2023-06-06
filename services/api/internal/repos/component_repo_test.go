package repos

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createComponent
func createComponent(ctx context.Context, t *testing.T, state repoTestState) *models.Component {
	app := createApp(ctx, t, state)
	componentID := domains.NewComponentID()

	component, err := state.componentRepo.Create(ctx, &models.Component{
		Name:  uuid.NewString(),
		AppID: app.ID,
		Model: models.Model{ID: componentID},
	})
	require.NoError(t, err)
	assert.NotNil(t, component)
	return component
}

func TestUpsertComponent(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should create an component successfully",
			fn: func(ctx context.Context, state repoTestState) {
				app := createApp(ctx, t, state)

				component, err := state.componentRepo.Create(ctx, &models.Component{
					Name:  uuid.NewString(),
					AppID: app.ID,
				})
				assert.NoError(t, err)
				assert.NotNil(t, component)
				assert.NotNil(t, component.ID)
			},
		},
		{
			desc: "should error when context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				component, err := state.componentRepo.Create(ctx, &models.Component{})
				assert.Error(t, err)
				assert.Nil(t, component)
			},
		},
	})
}

func TestDeleteComponent(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should delete an install successfully",
			fn: func(ctx context.Context, state repoTestState) {
				component := createComponent(ctx, t, state)

				_, err := state.componentRepo.Delete(ctx, component.ID)
				require.NoError(t, err)

				fetchedComponent, err := state.componentRepo.Get(ctx, component.ID)
				assert.Error(t, err)
				assert.Nil(t, fetchedComponent)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				componentID := domains.NewComponentID()
				deleted, err := state.componentRepo.Delete(ctx, componentID)
				assert.Error(t, err)
				assert.False(t, deleted)
			},
		},
	})
}

func TestGetComponent(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get an component successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origComponent := createComponent(ctx, t, state)

				component, err := state.componentRepo.Get(ctx, origComponent.ID)
				assert.NoError(t, err)
				assert.NotNil(t, component)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				component := createComponent(ctx, t, state)

				state.ctxCloseFn()
				fetchedComponent, err := state.componentRepo.Get(ctx, component.ID)
				assert.Error(t, err)
				assert.Nil(t, fetchedComponent)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				componentID := domains.NewComponentID()
				fetchedComponent, err := state.componentRepo.Get(ctx, componentID)
				assert.Error(t, err)
				assert.Nil(t, fetchedComponent)
			},
		},
	})
}

func TestComponentListByApp(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get all components successfully when no limit is set",
			fn: func(ctx context.Context, state repoTestState) {
				origComponent := createComponent(ctx, t, state)

				components, page, err := state.componentRepo.ListByApp(ctx, origComponent.AppID, &models.ConnectionOptions{})
				assert.Nil(t, err)
				assert.NotEmpty(t, page)
				assert.NotEmpty(t, components)

				// NOTE(jm): until we've fixed all bugs cleaning up all database objects from previous
				// runs, we can't guarantee this will be the only app in the list
				// assert.Equal(t, apps[0].ID, origApp.ID)
				// assert.Equals(t, len(apps), 1)
				found := false
				for _, component := range components {
					if component.ID == origComponent.ID {
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
				appID := domains.NewAppID()
				components, page, err := state.componentRepo.ListByApp(ctx, appID, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, components)
				assert.Nil(t, page)
			},
		},
	})
}
