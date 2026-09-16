package installcreate

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestResolveInputConfigUsesSelectedBranchConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	ctx := context.Background()
	want := &models.AppAppInputConfig{ID: "input-branch"}

	api.EXPECT().
		GetAppBranchAppConfigs(ctx, "app-id", "branch-id", gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, query *models.GetPaginatedQuery) ([]*models.AppAppConfig, bool, error) {
			require.Equal(t, branchConfigPageSize, query.Limit)
			return []*models.AppAppConfig{
				{
					ID:     "preview-config",
					Status: models.AppAppConfigStatusActive,
					Labels: models.GithubComNuoncoNuonPkgLabelsLabels{
						"source": string(models.AppAppBranchRunTypeGitDashPreviewDashRun),
					},
				},
				{ID: "inactive-config", Status: models.AppAppConfigStatusPending},
				{ID: "branch-config", Status: models.AppAppConfigStatusActive},
			}, false, nil
		})
	api.EXPECT().
		GetAppConfig(ctx, "app-id", "branch-config", gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, recurse *bool) (*models.AppAppConfig, error) {
			require.NotNil(t, recurse)
			require.True(t, *recurse)
			return &models.AppAppConfig{Input: want}, nil
		})

	got, err := ResolveInputConfig(ctx, api, "app-id", "branch-id")
	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestResolveInputConfigHydratesBranchInputGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	ctx := context.Background()
	group := &models.AppAppInputGroup{ID: "group-id"}
	input := &models.AppAppInput{ID: "input-id", GroupID: group.ID}

	api.EXPECT().
		GetAppBranchAppConfigs(ctx, "app-id", "branch-id", gomock.Any()).
		Return([]*models.AppAppConfig{
			{ID: "branch-config", Status: models.AppAppConfigStatusActive},
		}, false, nil)
	api.EXPECT().
		GetAppConfig(ctx, "app-id", "branch-config", gomock.Any()).
		Return(&models.AppAppConfig{
			Input: &models.AppAppInputConfig{
				InputGroups: []*models.AppAppInputGroup{group},
				Inputs:      []*models.AppAppInput{input},
			},
		}, nil)

	got, err := ResolveInputConfig(ctx, api, "app-id", "branch-id")
	require.NoError(t, err)
	require.Equal(t, []*models.AppAppInput{input}, got.InputGroups[0].AppInputs)
}

func TestResolveInputConfigUsesLegacyAppInputWithoutBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	ctx := context.Background()
	want := &models.AppAppInputConfig{ID: "input-latest"}

	api.EXPECT().GetAppInputLatestConfig(ctx, "app-id").Return(want, nil)

	got, err := ResolveInputConfig(ctx, api, "app-id", "")
	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestResolveInputConfigRejectsBranchWithoutEligibleConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	ctx := context.Background()

	api.EXPECT().
		GetAppBranchAppConfigs(ctx, "app-id", "branch-id", gomock.Any()).
		Return([]*models.AppAppConfig{
			{ID: "inactive-config", Status: models.AppAppConfigStatusPending},
		}, false, nil)

	_, err := ResolveInputConfig(ctx, api, "app-id", "branch-id")
	require.EqualError(t, err, "selected app branch branch-id has no active non-preview app config (no successful branch run to pin)")
}
