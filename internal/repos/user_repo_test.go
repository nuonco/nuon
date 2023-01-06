package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpsertUserOrg(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get an org's users successfully",
			fn: func(ctx context.Context, state repoTestState) {
				userID := uuid.NewString()
				org := createOrg(ctx, t, state.orgRepo)

				userOrg, err := state.userRepo.UpsertUserOrg(ctx, userID, org.ID)
				assert.Nil(t, err)
				assert.NotNil(t, userOrg)
			},
		},
		{
			desc: "should support upserts",
			fn: func(ctx context.Context, state repoTestState) {
				userID := uuid.NewString()
				org := createOrg(ctx, t, state.orgRepo)

				userOrg, err := state.userRepo.UpsertUserOrg(ctx, userID, org.ID)
				assert.Nil(t, err)
				assert.NotNil(t, userOrg)
			},
		},
		{
			desc: "should error with a context canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				userOrg, err := state.userRepo.UpsertUserOrg(ctx, uuid.NewString(), uuid.New())
				assert.NotNil(t, err)
				assert.Nil(t, userOrg)
			},
		},
	})
}
