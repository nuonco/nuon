package repos

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/stretchr/testify/assert"
)

func TestUpsertUserOrg(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

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
				orgID := domains.NewOrgID()
				userOrg, err := state.userRepo.UpsertUserOrg(ctx, uuid.NewString(), orgID)
				assert.NotNil(t, err)
				assert.Nil(t, userOrg)
			},
		},
	})
}
