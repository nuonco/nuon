package repos

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func createOrg(ctx context.Context, t *testing.T, orgRepo OrgRepo) *models.Org {
	id := domains.NewOrgID()
	org, err := orgRepo.Create(ctx, &models.Org{
		Name:        uuid.NewString(),
		CreatedByID: uuid.NewString(),
		Model:       models.Model{ID: id},
	})
	assert.Nil(t, err)
	return org
}

func TestUpsertOrg(t *testing.T) {
	id := domains.NewOrgID()
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should create org successfully",
			fn: func(ctx context.Context, state repoTestState) {
				orgInput := models.Org{
					Model:           models.Model{ID: id},
					Name:            uuid.NewString(),
					GithubInstallID: fmt.Sprintf("%d", faker.New().UInt32()),
				}

				org, err := state.orgRepo.Create(ctx, &orgInput)
				assert.Nil(t, err)
				assert.NotNil(t, org)
				assert.NotNil(t, org.ID)
			},
		},
		{
			desc: "should set isNew properly for new orgs",
			fn: func(ctx context.Context, state repoTestState) {
				orgInput := models.Org{
					Name: uuid.NewString(),
				}
				org, err := state.orgRepo.Create(ctx, &orgInput)
				assert.Nil(t, err)
				assert.NotNil(t, org)
				assert.NotNil(t, org.ID)
				assert.True(t, org.IsNew)

				org, err = state.orgRepo.Get(ctx, org.ID)
				assert.NoError(t, err)
				assert.False(t, org.IsNew)
			},
		},
		{
			desc: "should update an org successfully",
			fn: func(ctx context.Context, state repoTestState) {
				org := createOrg(ctx, t, state.orgRepo)
				org.Name += "abc"
				org.GithubInstallID = fmt.Sprintf("%d", faker.New().UInt32())
				org, err := state.orgRepo.Update(ctx, org)
				assert.Nil(t, err)

				fetchedOrg, err := state.orgRepo.Get(ctx, org.ID)
				assert.Nil(t, err)
				assert.Equal(t, org.Name, fetchedOrg.Name)
				assert.Equal(t, org.GithubInstallID, fetchedOrg.GithubInstallID)
			},
		},
		{
			desc: "should error when context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				_, err := state.orgRepo.Create(ctx, &models.Org{
					Name: uuid.NewString(),
				})
				assert.NotNil(t, err)
			},
		},
	})
}

func TestDeleteOrg(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should delete an org successfully",
			fn: func(ctx context.Context, state repoTestState) {
				org := createOrg(ctx, t, state.orgRepo)
				deleted, err := state.orgRepo.Delete(ctx, org.ID)
				assert.True(t, deleted)
				assert.Nil(t, err)

				org, err = state.orgRepo.Get(ctx, org.ID)
				assert.NotNil(t, err)
				assert.Nil(t, org)
			},
		},
		{
			desc: "should error if context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				org := createOrg(ctx, t, state.orgRepo)
				state.ctxCloseFn()

				deleted, err := state.orgRepo.Delete(ctx, org.ID)
				assert.NotNil(t, err)
				assert.False(t, deleted)
			},
		},
		{
			desc: "should return false if not found",
			fn: func(ctx context.Context, state repoTestState) {
				orgID := domains.NewOrgID()
				deleted, err := state.orgRepo.Delete(ctx, orgID)
				assert.Nil(t, err)
				assert.False(t, deleted)
			},
		},
	})
}

func TestGetOrg(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should get an org successfully",
			fn: func(ctx context.Context, state repoTestState) {
				org := createOrg(ctx, t, state.orgRepo)
				org, err := state.orgRepo.Get(ctx, org.ID)
				assert.Nil(t, err)
				assert.NotNil(t, org)
			},
		},
		{
			desc: "should error if context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				org := createOrg(ctx, t, state.orgRepo)
				state.ctxCloseFn()
				org, err := state.orgRepo.Get(ctx, org.ID)
				assert.NotNil(t, err)
				assert.Nil(t, org)
			},
		},
		{
			desc: "should return an error if not found",
			fn: func(ctx context.Context, state repoTestState) {
				orgID := domains.NewOrgID()
				org, err := state.orgRepo.Get(ctx, orgID)
				assert.Nil(t, org)
				assert.NotNil(t, err)
			},
		},
	})
}

/*
func TestOrgGetPageByUser(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get all orgs a user is a part of",
			fn: func(ctx context.Context, state repoTestState) {
				origUser := createUser(ctx, t, state, true)

				orgs, page, err := state.orgRepo.GetPageByUser(ctx, origUser.ID, &models.ConnectionOptions{})
				assert.Nil(t, err)
				assert.NotEmpty(t, page)
				assert.NotEmpty(t, orgs)

				assert.Equal(t, orgs[0].ID, origUser.Orgs[0].ID)
			},
		},
		{
			desc: "should error with a context canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				orgs, page, err := state.orgRepo.GetPageByUser(ctx, uuid.Nil, &models.ConnectionOptions{})
				assert.NotNil(t, err)
				assert.Nil(t, orgs)
				assert.Nil(t, page)
			},
		},
	})
}
*/
