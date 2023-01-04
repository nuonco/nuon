package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/stretchr/testify/assert"
)

// createUser
func createUser(ctx context.Context, t *testing.T, state repoTestState, createOrg bool) *models.User {
	user, err := state.userRepo.Upsert(ctx, &models.User{
		Email:     fkr.Internet().Email(),
		FirstName: fkr.Person().FirstName(),
		LastName:  fkr.Person().LastName(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, user)
	if !createOrg {
		return user
	}

	org, err := state.orgRepo.Create(ctx, &models.Org{
		Name: uuid.NewString(),
		Slug: uuid.NewString(),
	})
	assert.Nil(t, err)
	_, err = state.userRepo.UpsertUserOrg(ctx, user.ID, org.ID)
	assert.NoError(t, err)
	user.Orgs = []models.Org{*org}
	return user
}

func TestUpsertUser(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should create a user successfully",
			fn: func(ctx context.Context, state repoTestState) {
				userInput := &models.User{
					Email:     fkr.Internet().Email(),
					FirstName: fkr.Person().FirstName(),
					LastName:  fkr.Person().LastName(),
				}
				user, err := state.userRepo.Upsert(ctx, userInput)
				assert.Nil(t, err)
				assert.NotNil(t, user)
				assert.NotNil(t, user.ID)
			},
		},
		{
			desc: "should upsert when creating with dupe email",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, false)

				user, err := state.userRepo.Upsert(ctx, &models.User{
					Email:     origUser.Email,
					FirstName: fkr.Person().FirstName(),
					LastName:  fkr.Person().LastName(),
				})
				assert.NotNil(t, user)
				assert.Nil(t, err)
				assert.Equal(t, origUser.ID, user.ID)
			},
		},
		{
			desc: "should create when creating with dupe name",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, false)

				user, err := state.userRepo.Upsert(ctx, &models.User{
					Email:     fkr.Internet().Email(),
					FirstName: origUser.FirstName,
					LastName:  origUser.LastName,
				})
				assert.NotNil(t, user)
				assert.Nil(t, err)
				assert.NotEqual(t, origUser.ID, user.ID)
			},
		},
		{
			desc: "should error when context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				_, err := state.userRepo.Upsert(ctx, &models.User{
					Email:     fkr.Internet().Email(),
					FirstName: fkr.Person().FirstName(),
					LastName:  fkr.Person().LastName(),
				})
				assert.NotNil(t, err)
			},
		},
	})
}

func TestDeleteUser(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should delete a user successfully",
			fn: func(ctx context.Context, state repoTestState) {
				user := createUser(ctx, t, state, false)

				deleted, err := state.userRepo.Delete(ctx, user.ID)
				assert.Nil(t, err)
				assert.True(t, deleted)

				fetchedUser, err := state.userRepo.Get(ctx, user.ID)
				assert.NotNil(t, err)
				assert.Nil(t, fetchedUser)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				deleted, err := state.userRepo.Delete(ctx, uuid.New())
				assert.False(t, deleted)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetUser(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get a user successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, false)

				user, err := state.userRepo.Get(ctx, origUser.ID)
				assert.Nil(t, err)
				assert.NotNil(t, user)
			},
		},
		{
			desc: "should get a user with orgs successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, true)

				user, err := state.userRepo.Get(ctx, origUser.ID)
				assert.Nil(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.Orgs)
				assert.Equal(t, user.Orgs[0].ID, origUser.Orgs[0].ID)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				user := createUser(ctx, t, state, true)

				state.ctxCloseFn()
				fetchedUser, err := state.userRepo.Get(ctx, user.ID)
				assert.Nil(t, fetchedUser)
				assert.NotNil(t, err)

				fetchedUser, err = state.userRepo.GetByEmail(ctx, user.Email)
				assert.Nil(t, fetchedUser)
				assert.NotNil(t, err)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedUser, err := state.userRepo.Get(ctx, uuid.New())
				assert.Nil(t, fetchedUser)
				assert.NotNil(t, err)

				fetchedUser, err = state.userRepo.GetByEmail(ctx, fkr.Internet().Email())
				assert.Nil(t, fetchedUser)
				assert.NotNil(t, err)
			},
		},
	})
}

func TestGetUsersByOrg(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get an org's users successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, true)

				users, err := state.userRepo.GetByOrg(ctx, origUser.Orgs[0].ID)
				assert.Nil(t, err)
				assert.NotEmpty(t, users)
				assert.Equal(t, users[0].ID, origUser.ID)
			},
		},
		{
			desc: "should error with a context canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				users, err := state.userRepo.GetByOrg(ctx, uuid.New())
				assert.NotNil(t, err)
				assert.Nil(t, users)
			},
		},
	})
}

func TestUserGetPageAll(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get all users successfully when no limit is set",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, false)

				users, page, err := state.userRepo.GetPageAll(ctx, &models.ConnectionOptions{})
				assert.Nil(t, err)
				assert.NotEmpty(t, page)
				assert.NotEmpty(t, users)

				// NOTE(jm): until we've fixed all bugs cleaning up all database objects from previous
				// runs, we can't guarantee this will be the only user in the list
				// assert.Equal(t, users[0].ID, origUser.ID)
				// assert.Equals(t, len(users), 1)
				found := false
				for _, user := range users {
					if user.ID == origUser.ID {
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
				users, page, err := state.userRepo.GetPageAll(ctx, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, users)
				assert.Nil(t, page)
			},
		},
	})
}

func TestUpsertUserOrg(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get an org's users successfully",
			fn: func(ctx context.Context, state repoTestState) {
				user := createUser(ctx, t, state, false)
				org := createOrg(ctx, t, state.orgRepo)

				userOrg, err := state.userRepo.UpsertUserOrg(ctx, user.ID, org.ID)
				assert.Nil(t, err)
				assert.NotNil(t, userOrg)

				fetchedUser, err := state.userRepo.Get(ctx, user.ID)
				assert.Nil(t, err)
				assert.NotEmpty(t, fetchedUser.Orgs)
				assert.Equal(t, fetchedUser.Orgs[0].ID, org.ID)
			},
		},
		{
			desc: "should support upserts",
			fn: func(ctx context.Context, state repoTestState) {
				user := createUser(ctx, t, state, false)
				org := createOrg(ctx, t, state.orgRepo)

				userOrg, err := state.userRepo.UpsertUserOrg(ctx, user.ID, org.ID)
				assert.Nil(t, err)
				assert.NotNil(t, userOrg)
			},
		},
		{
			desc: "should error with a context canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				userOrg, err := state.userRepo.UpsertUserOrg(ctx, uuid.New(), uuid.New())
				assert.NotNil(t, err)
				assert.Nil(t, userOrg)
			},
		},
	})
}
