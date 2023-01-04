package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/stretchr/testify/assert"
)

// createApp
func createApp(ctx context.Context, t *testing.T, state repoTestState) *models.App {
	user := createUser(ctx, t, state, true)

	app, err := state.appRepo.Upsert(ctx, &models.App{
		Name:        fkr.App().Name(),
		CreatedByID: user.ID,
		OrgID:       user.Orgs[0].ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, app)
	return app
}

func TestUpsertApp(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should create an app successfully",
			fn: func(ctx context.Context, state repoTestState) {
				user := createUser(ctx, t, state, true)

				appInput := &models.App{
					Name:        fkr.App().Name(),
					CreatedByID: user.ID,
					OrgID:       user.Orgs[0].ID,
				}
				app, err := state.appRepo.Upsert(ctx, appInput)
				assert.Nil(t, err)
				assert.NotNil(t, app)
				assert.NotNil(t, app.ID)
			},
		},
		{
			desc: "should upsert when creating with same ID",
			fn: func(ctx context.Context, state repoTestState) {
				origApp := createApp(ctx, t, state)
				appInput := &models.App{
					Model: models.Model{ID: origApp.ID},
					Name:  origApp.Name + "a",
				}
				app, err := state.appRepo.Upsert(ctx, appInput)
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
				app, err := state.appRepo.Upsert(ctx, &models.App{})
				assert.NotNil(t, err)
				assert.Nil(t, app)
			},
		},
	})
}

func TestDeleteApp(t *testing.T) {
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
				state.ctxCloseFn()
				deleted, err := state.appRepo.Delete(ctx, uuid.New())
				assert.False(t, deleted)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetApp(t *testing.T) {
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
				fetchedApp, err := state.appRepo.Get(ctx, uuid.New())
				assert.Nil(t, fetchedApp)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetAppBySlug(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get an app successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origApp := createApp(ctx, t, state)

				app, err := state.appRepo.GetBySlug(ctx, origApp.Slug)
				assert.Nil(t, err)
				assert.NotNil(t, app)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				app := createApp(ctx, t, state)

				state.ctxCloseFn()
				fetchedApp, err := state.appRepo.GetBySlug(ctx, app.Slug)
				assert.Nil(t, fetchedApp)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedApp, err := state.appRepo.GetBySlug(ctx, "abc")
				assert.Nil(t, fetchedApp)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestAppGetPageAll(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get all apps successfully when no limit is set",
			fn: func(ctx context.Context, state repoTestState) {
				origApp := createApp(ctx, t, state)

				apps, page, err := state.appRepo.GetPageAll(ctx, &models.ConnectionOptions{})
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
				apps, page, err := state.appRepo.GetPageAll(ctx, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, apps)
				assert.Nil(t, page)
			},
		},
	})
}

func TestAppGetPageByOrg(t *testing.T) {
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
				apps, page, err := state.appRepo.GetPageByOrg(ctx, uuid.New(), &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, apps)
				assert.Nil(t, page)
			},
		},
	})
}
