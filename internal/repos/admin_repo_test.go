package repos

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
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

func TestGetSandboxVersion(t *testing.T) {
	execRepoTests(t, []repoTest{
		{
			desc: "should get a sandbox version successfully",
			fn: func(ctx context.Context, state repoTestState) {
				origSandboxVersion := createSandboxVersion(ctx, t, state)

				sandboxVersion, err := state.adminRepo.GetSandboxVersion(ctx, origSandboxVersion.ID)
				assert.NoError(t, err)
				assert.NotNil(t, sandboxVersion)
			},
		},
		{
			desc: "should error with canceled context",
			fn: func(ctx context.Context, state repoTestState) {
				sandboxVersion := createSandboxVersion(ctx, t, state)

				state.ctxCloseFn()
				fetchedSandboxVersion, err := state.adminRepo.GetSandboxVersion(ctx, sandboxVersion.ID)
				assert.Error(t, err)
				assert.Nil(t, fetchedSandboxVersion)
			},
		},
		{
			desc: "should error with not found",
			fn: func(ctx context.Context, state repoTestState) {
				fetchedSandboxVersion, err := state.adminRepo.GetSandboxVersion(ctx, uuid.New())
				assert.Error(t, err)
				assert.Nil(t, fetchedSandboxVersion)
			},
		},
	})
}

func TestUpsertSandboxVersion(t *testing.T) {
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
				sandboxVersion.ID = uuid.New()
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
