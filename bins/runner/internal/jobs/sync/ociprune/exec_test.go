package ociprune

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/oci/prune"
)

type fakeRegistry struct {
	manifests []prune.Manifest
	deleted   [][]prune.Deletion
}

func (f *fakeRegistry) RepositoryURL() string { return "123.dkr.ecr.us-east-1.amazonaws.com/inst" }

func (f *fakeRegistry) ListManifests(ctx context.Context) ([]prune.Manifest, error) {
	return f.manifests, nil
}

func (f *fakeRegistry) Delete(ctx context.Context, deletions []prune.Deletion) ([]string, []prune.DeleteFailure, error) {
	f.deleted = append(f.deleted, deletions)
	var out []string
	for _, d := range deletions {
		out = append(out, d.Digest)
	}
	return out, nil, nil
}

func newTestHandler(t *testing.T, dryRun bool, reg *fakeRegistry, scan scanClusterFn) (*handler, *nuonrunner.MockClient) {
	ctrl := gomock.NewController(t)
	client := nuonrunner.NewMockClient(ctrl)
	h := &handler{
		apiClient: client,
		openRegistry: func(ctx context.Context, cfg *configs.OCIRegistryRepository) (registry, error) {
			return reg, nil
		},
		scanCluster: scan,
		state: &handlerState{
			jobID:          "job",
			jobExecutionID: "exec",
			plan: &plantypes.PruneOCIPlan{
				Registry:    &configs.OCIRegistryRepository{},
				DryRun:      dryRun,
				OwnedTags:   []string{"bld-old", "bld-new"},
				MaxDeletes:  200,
				ClusterInfo: &kube.ClusterInfo{},
			},
		},
	}
	return h, client
}

func oldManifests() []prune.Manifest {
	old := time.Now().Add(-90 * 24 * time.Hour)
	return []prune.Manifest{
		{Digest: "sha256:old", Tags: []string{"bld-old"}, PushedAt: old},
		{Digest: "sha256:new", Tags: []string{"bld-new"}, PushedAt: old},
	}
}

func testCtx() context.Context {
	return pkgctx.SetLogger(context.Background(), zap.NewNop())
}

func TestExecClusterScanFailureDeletesNothing(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		reg := &fakeRegistry{manifests: oldManifests()}
		scanErr := errors.New("Unauthorized")
		h, client := newTestHandler(t, dryRun, reg, func(ctx context.Context, ci *kube.ClusterInfo, repo string) (*prune.InUse, error) {
			return nil, scanErr
		})

		var got *models.ServiceCreateRunnerJobExecutionResultRequest
		client.EXPECT().
			CreateJobExecutionResult(gomock.Any(), "job", "exec", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
				got = req
				return &models.AppRunnerJobExecutionResult{}, nil
			})

		err := h.Exec(testCtx(), &models.AppRunnerJob{ID: "job"}, &models.AppRunnerJobExecution{ID: "exec"})

		require.ErrorIs(t, err, scanErr)
		assert.Empty(t, reg.deleted, "dry_run=%t", dryRun)
		assert.Nil(t, h.state.result)
		require.NotNil(t, got)
		assert.False(t, got.Success)
		assert.Equal(t, "scan cluster", got.ErrorMetadata["step"])
	}
}

func TestExecClusterScanSuccessDeletesUnused(t *testing.T) {
	reg := &fakeRegistry{manifests: oldManifests()}
	h, client := newTestHandler(t, false, reg, func(ctx context.Context, ci *kube.ClusterInfo, repo string) (*prune.InUse, error) {
		return &prune.InUse{Tags: []string{"bld-new"}}, nil
	})

	var got *models.ServiceCreateRunnerJobExecutionResultRequest
	client.EXPECT().
		CreateJobExecutionResult(gomock.Any(), "job", "exec", gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
			got = req
			return &models.AppRunnerJobExecutionResult{}, nil
		})

	err := h.Exec(testCtx(), &models.AppRunnerJob{ID: "job"}, &models.AppRunnerJobExecution{ID: "exec"})

	require.NoError(t, err)
	require.Len(t, reg.deleted, 1)
	assert.Equal(t, []prune.Deletion{{Digest: "sha256:old", Tags: []string{"bld-old"}}}, reg.deleted[0])

	require.NotNil(t, got)
	assert.True(t, got.Success)
	var res Result
	require.NoError(t, json.Unmarshal([]byte(got.Contents), &res))
	assert.Equal(t, OutcomePruned, res.Outcome)
	assert.Equal(t, []string{"sha256:old"}, res.Deleted)
	assert.Equal(t, 1, res.KeptByWhy[prune.KeepReasonInUse])
}
