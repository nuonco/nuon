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

// createApp
func createApp(ctx context.Context, t *testing.T, state repoTestState) *models.App {
	userID := uuid.NewString()
	org := createOrg(ctx, t, state.orgRepo)
	appID := domains.NewAppID()
	app, err := state.appRepo.Create(ctx, &models.App{
		Name:        uuid.NewString(),
		CreatedByID: userID,
		OrgID:       org.ID,
		Model:       models.Model{ID: appID},
	})
	assert.Nil(t, err)
	assert.NotNil(t, app)
	return app
}

func TestUpsertApp(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should create an app successfully",
			fn: func(ctx context.Context, state repoTestState) {
				userID := uuid.NewString()
				org := createOrg(ctx, t, state.orgRepo)
				appID := domains.NewAppID()

				appInput := &models.App{
					Name:        uuid.NewString(),
					CreatedByID: userID,
					OrgID:       org.ID,
					Model:       models.Model{ID: appID},
				}
				app, err := state.appRepo.Create(ctx, appInput)
				assert.Nil(t, err)
				assert.NotNil(t, app)
				assert.NotNil(t, app.ID)
			},
		},
		{
			desc: "should update an app successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origApp := createApp(ctx, t, state)
				appInput := &models.App{
					Model: models.Model{ID: origApp.ID},
					Name:  origApp.Name + "a",
				}
				app, err := state.appRepo.Update(ctx, appInput)
				assert.Nil(t, err)
				assert.NotNil(t, app)

				app, err = state.appRepo.Get(ctx, origApp.ID)
				assert.Nil(t, err)
				assert.Equal(t, app.Name, origApp.Name+"a")
			},
		},
		{
			desc: "should error when context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				app, err := state.appRepo.Create(ctx, &models.App{})
				assert.NotNil(t, err)
				assert.Nil(t, app)
			},
		},
	})
}

func TestDeleteApp(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should delete an app successfully",
			fn: func(ctx context.Context, state repoTestState) {
				app := createApp(ctx, t, state)

				_, err := state.appRepo.Delete(ctx, app.ID)
				assert.Nil(t, err)

				fetchedApp, err := state.appRepo.Get(ctx, app.ID)
				assert.NotNil(t, err)
				assert.Nil(t, fetchedApp)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				appID := domains.NewAppID()
				state.ctxCloseFn()
				deleted, err := state.appRepo.Delete(ctx, appID)
				assert.False(t, deleted)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetApp(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get an app successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origApp := createApp(ctx, t, state)

				app, err := state.appRepo.Get(ctx, origApp.ID)
				assert.Nil(t, err)
				assert.NotNil(t, app)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				app := createApp(ctx, t, state)

				state.ctxCloseFn()
				fetchedApp, err := state.appRepo.Get(ctx, app.ID)
				assert.Nil(t, fetchedApp)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				appID := domains.NewAppID()
				fetchedApp, err := state.appRepo.Get(ctx, appID)
				assert.Nil(t, fetchedApp)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestAppGetPageByOrg(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get all apps successfully when no limit is set",
			fn: func(ctx context.Context, state repoTestState) {
				origApp := createApp(ctx, t, state)

				apps, page, err := state.appRepo.GetPageByOrg(ctx, origApp.OrgID, &models.ConnectionOptions{})
				assert.Nil(t, err)
				assert.NotEmpty(t, page)
				assert.NotEmpty(t, apps)

				// NOTE(jm): until we've fixed all bugs cleaning up all database objects from previous
				// runs, we can't guarantee this will be the only app in the list
				// assert.Equal(t, apps[0].ID, origApp.ID)
				// assert.Equals(t, len(apps), 1)
				found := false
				for _, app := range apps {
					if app.ID == origApp.ID {
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
				orgID := domains.NewOrgID()
				apps, page, err := state.appRepo.GetPageByOrg(ctx, orgID, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, apps)
				assert.Nil(t, page)
			},
		},
	})
}
