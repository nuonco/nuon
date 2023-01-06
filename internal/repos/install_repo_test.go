package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/stretchr/testify/assert"
)

// createInstall
func createInstall(ctx context.Context, t *testing.T, state repoTestState) *models.Install {
	app := createApp(ctx, t, state)
	userID := uuid.NewString()

	install, err := state.installRepo.Create(ctx, &models.Install{
		Name:        uuid.NewString(),
		CreatedByID: userID,
		AppID:       app.ID,
		AWSSettings: &models.AWSSettings{
			Region: models.AWSRegionUsEast1,
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, install)
	return install
}

func TestUpsertInstall(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should create an install successfully",
			fn: func(ctx context.Context, state repoTestState) {
				app := createApp(ctx, t, state)
				userID := uuid.NewString()

				installInput := &models.Install{
					Name:        uuid.NewString(),
					CreatedByID: userID,
					AppID:       app.ID,
				}
				install, err := state.installRepo.Create(ctx, installInput)
				assert.Nil(t, err)
				assert.NotNil(t, install)
				assert.NotNil(t, install.ID)
			},
		},
		{
			desc: "should error when context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				install, err := state.installRepo.Create(ctx, &models.Install{})
				assert.NotNil(t, err)
				assert.Nil(t, install)
			},
		},
	})
}

func TestDeleteInstall(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should delete an install successfully",
			fn: func(ctx context.Context, state repoTestState) {
				install := createInstall(ctx, t, state)

				_, err := state.installRepo.Delete(ctx, install.ID)
				assert.Nil(t, err)

				fetchedInstall, err := state.installRepo.Get(ctx, install.ID)
				assert.NotNil(t, err)
				assert.Nil(t, fetchedInstall)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				deleted, err := state.installRepo.Delete(ctx, uuid.New())
				assert.False(t, deleted)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetInstall(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get an install successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origInstall := createInstall(ctx, t, state)

				install, err := state.installRepo.Get(ctx, origInstall.ID)
				assert.Nil(t, err)
				assert.NotNil(t, install)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				install := createInstall(ctx, t, state)

				state.ctxCloseFn()
				fetchedInstall, err := state.installRepo.Get(ctx, install.ID)
				assert.Nil(t, fetchedInstall)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedInstall, err := state.installRepo.Get(ctx, uuid.New())
				assert.Nil(t, fetchedInstall)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should get an AWSSettings should not be nil",
			fn: func(ctx context.Context, state repoTestState) {
				origInstall := createInstall(ctx, t, state)

				install, err := state.installRepo.Get(ctx, origInstall.ID)
				assert.Nil(t, err)
				assert.NotNil(t, install)
				assert.NotNil(t, install.Settings)
			},
		},
	})
}

func TestInstallListByApp(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get all apps successfully when no limit is set",
			fn: func(ctx context.Context, state repoTestState) {
				origInstall := createInstall(ctx, t, state)

				installs, page, err := state.installRepo.ListByApp(ctx, origInstall.AppID, &models.ConnectionOptions{})
				assert.Nil(t, err)
				assert.NotEmpty(t, page)
				assert.NotEmpty(t, installs)

				// NOTE(jm): until we've fixed all bugs cleaning up all database objects from previous
				// runs, we can't guarantee this will be the only app in the list
				// assert.Equal(t, apps[0].ID, origApp.ID)
				// assert.Equals(t, len(apps), 1)
				found := false
				for _, install := range installs {
					if install.ID == origInstall.ID {
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
				installs, page, err := state.installRepo.ListByApp(ctx, uuid.New(), &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, installs)
				assert.Nil(t, page)
			},
		},
	})
}
