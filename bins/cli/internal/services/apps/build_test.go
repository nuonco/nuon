package apps

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type noAppConfigsAPI struct {
	nuon.Client
	appSyncDisabled bool
}

func (a *noAppConfigsAPI) GetApp(_ context.Context, appID string) (*models.AppApp, error) {
	return &models.AppApp{ID: appID}, nil
}

func (a *noAppConfigsAPI) GetAppConfigs(context.Context, string, *models.GetPaginatedQuery) ([]*models.AppAppConfig, bool, error) {
	return nil, false, nil
}

func (a *noAppConfigsAPI) GetOrg(context.Context) (*models.AppOrg, error) {
	return &models.AppOrg{Features: models.TypesStringBoolMap{"disable-app-sync": a.appSyncDisabled}}, nil
}

func TestBuildWithoutAppConfigs(t *testing.T) {
	tests := map[string]struct {
		appSyncDisabled bool
		wantErr         string
	}{
		"app sync enabled":  {wantErr: "no app configs found; run `nuon apps sync` or run an app branch first"},
		"app sync disabled": {appSyncDisabled: true, wantErr: "no app configs found; run an app branch first"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := &Service{api: &noAppConfigsAPI{appSyncDisabled: tc.appSyncDisabled}}

			err := s.Build(context.Background(), "app-1", "", "")
			require.EqualError(t, err, tc.wantErr)
		})
	}
}
