package creator

import (
	"context"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestCheckInstallNameFindsExactMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	ctx := context.Background()

	api.EXPECT().
		GetAppInstalls(ctx, "app-id", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
			require.Equal(t, "staging", query.Q)
			return []*models.AppInstall{
				{Name: "staging-copy"},
				{Name: "staging"},
			}, false, nil
		})

	msg := checkInstallNameCmd(model{ctx: ctx, api: api, appID: "app-id"}, "staging", 7)()
	result := msg.(nameCheckedMsg)
	require.True(t, result.exists)
	require.Equal(t, 7, result.generation)
}

func TestValidateNameRejectsDuplicate(t *testing.T) {
	input := textinput.New()
	input.SetValue("staging")
	m := model{
		inputs:            []textinput.Model{input},
		nameChecked:       "staging",
		nameValidationErr: errDuplicateInstallName("staging"),
	}

	require.EqualError(t, m.validateName(), `an install named "staging" already exists`)
}

func TestValidateNameBlocksWhileCheckIsPendingOrStale(t *testing.T) {
	input := textinput.New()
	input.SetValue("staging")

	m := model{inputs: []textinput.Model{input}, nameChecking: true}
	require.EqualError(t, m.validateName(), `checking whether install name "staging" is available`)

	m.nameChecking = false
	m.nameChecked = "old-name"
	require.EqualError(t, m.validateName(), `checking whether install name "staging" is available`)
}

func TestCheckInstallNamePaginatesUntilExactMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	ctx := context.Background()
	first := api.EXPECT().
		GetAppInstalls(ctx, "app-id", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
			require.Zero(t, query.Offset)
			return []*models.AppInstall{{Name: "staging-copy"}}, true, nil
		})
	api.EXPECT().
		GetAppInstalls(ctx, "app-id", gomock.Any()).
		After(first).
		DoAndReturn(func(_ context.Context, _ string, query *models.GetPaginatedQuery) ([]*models.AppInstall, bool, error) {
			require.Equal(t, 1, query.Offset)
			return []*models.AppInstall{{Name: "staging"}}, false, nil
		})

	msg := checkInstallNameCmd(model{ctx: ctx, api: api, appID: "app-id"}, "staging", 8)()
	result := msg.(nameCheckedMsg)
	require.True(t, result.exists)
	require.Equal(t, 8, result.generation)
}

func TestUpdateDropsStaleNameCheckResult(t *testing.T) {
	m := model{
		nameCheckGeneration: 2,
		nameChecking:        true,
	}

	updated, _ := m.Update(nameCheckedMsg{
		name:       "staging",
		generation: 1,
		exists:     true,
	})
	got := updated.(model)
	require.True(t, got.nameChecking)
	require.NoError(t, got.nameValidationErr)
}
