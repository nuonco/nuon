package repos

import (
	"context"
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSandboxVersion(ctx context.Context, t *testing.T, state repoTestState) *models.SandboxVersion {
	sandboxVersion, err := state.adminRepo.UpsertSandboxVersion(ctx, &models.SandboxVersion{
		SandboxName:    "sandbox-name",
		SandboxVersion: "1.0",
		TfVersion:      "10.1",
	})
	require.NoError(t, err)
	assert.NotNil(t, sandboxVersion)
	return sandboxVersion
}

func TestGetSandboxVersionByID(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	id := domains.NewSandboxID()
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should return error if record does not exist",
			fn: func(ctx context.Context, state repoTestState) {
				sandboxVersion, err := state.adminRepo.GetSandboxVersionByID(ctx, id)
				assert.Error(t, err)
				assert.Nil(t, sandboxVersion)
			},
		},
		{
			desc: "should fetch the record if it exists",
			fn: func(ctx context.Context, state repoTestState) {
				origSandboxVersion := createSandboxVersion(ctx, t, state)

				sandboxVersion, err := state.adminRepo.GetSandboxVersionByID(ctx, origSandboxVersion.ID)
				assert.NoError(t, err)
				assert.NotNil(t, sandboxVersion)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				fetchedSandboxVersion, err := state.adminRepo.GetSandboxVersionByID(ctx, id)
				assert.Error(t, err)
				assert.Nil(t, fetchedSandboxVersion)
			},
		},
	})
}

func TestGetLatestSandboxVersion(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should fetch the most recent sandbox version",
			fn: func(ctx context.Context, state repoTestState) {
				_ = createSandboxVersion(ctx, t, state)
				_ = createSandboxVersion(ctx, t, state)
				latestSandboxVersion := createSandboxVersion(ctx, t, state)
				fetchedSandboxVersion, err := state.adminRepo.GetLatestSandboxVersion(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, fetchedSandboxVersion)
				assert.Equal(t, latestSandboxVersion.ID, fetchedSandboxVersion.ID)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				fetchedSandboxVersion, err := state.adminRepo.GetLatestSandboxVersion(ctx)
				assert.Error(t, err)
				assert.Nil(t, fetchedSandboxVersion)
			},
		},
	})
}

func TestUpsertSandboxVersion(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	execRepoTests(t, []repoTest{
		{
			desc: "should create a sandbox version successfully",
			fn: func(ctx context.Context, state repoTestState) {
				sandboxVersion := createSandboxVersion(ctx, t, state)
				assert.NotNil(t, sandboxVersion.ID)
			},
		},
		{
			desc: "should update a sandbox version successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origSandboxVersion := createSandboxVersion(ctx, t, state)
				origSandboxVersion.SandboxVersion = "2.0"
				updatedSandboxVersion, err := state.adminRepo.UpsertSandboxVersion(ctx, origSandboxVersion)
				require.NoError(t, err)
				assert.NotNil(t, updatedSandboxVersion)
				assert.Equal(t, updatedSandboxVersion.SandboxVersion, origSandboxVersion.SandboxVersion)
			},
		},
		{
			desc: "should error out if id is provided but not found",
			fn: func(ctx context.Context, state repoTestState) {
				sandboxVersion := &models.SandboxVersion{
					SandboxName:    "sandbox-name",
					SandboxVersion: "1.0",
					TfVersion:      "10.1",
				}
				id := domains.NewSandboxID()
				sandboxVersion.ID = id
				updatedSandboxVersion, err := state.adminRepo.UpsertSandboxVersion(ctx, sandboxVersion)
				require.Error(t, err)
				assert.Nil(t, updatedSandboxVersion)
			},
		},
		{
			desc: "should error when context is canceled",
			fn: func(ctx context.Context, state repoTestState) {
				state.ctxCloseFn()
				sandboxVersion, err := state.adminRepo.UpsertSandboxVersion(ctx, &models.SandboxVersion{})
				assert.Error(t, err)
				assert.Nil(t, sandboxVersion)
			},
		},
	})
}
