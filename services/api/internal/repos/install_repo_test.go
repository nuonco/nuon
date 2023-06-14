package repos

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
)

// createInstall
func createInstall(ctx context.Context, t *testing.T, state repoTestState) *models.Install {
	app := createApp(ctx, t, state)
	userID := uuid.NewString()
	installID := domains.NewInstallID()

	install, err := state.installRepo.Create(ctx, &models.Install{
		Model:       models.Model{ID: installID},
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
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should create an install successfully",
			fn: func(ctx context.Context, state repoTestState) {
				app := createApp(ctx, t, state)
				userID := uuid.NewString()
				installID := domains.NewInstallID()

				installInput := &models.Install{
					Model:       models.Model{ID: installID},
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
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

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
				installID := domains.NewInstallID()
				deleted, err := state.installRepo.Delete(ctx, installID)
				assert.False(t, deleted)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetInstall(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

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
				installID := domains.NewInstallID()
				fetchedInstall, err := state.installRepo.Get(ctx, installID)
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
				assert.NotNil(t, install.AWSSettings)
			},
		},
	})
}

func TestInstallListByApp(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

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
				appID := domains.NewAppID()
				installs, page, err := state.installRepo.ListByApp(ctx, appID, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, installs)
				assert.Nil(t, page)
			},
		},
	})
}
