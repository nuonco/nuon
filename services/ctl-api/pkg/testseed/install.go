package testseed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *Seeder) EnsureInstall(ctx context.Context, t *testing.T, appID string) *app.Install {
	org, err := cctx.OrgFromContext(ctx)
	require.NoError(t, err, "context must have org set via EnsureOrg")

	accountID, err := cctx.AccountIDFromContext(ctx)
	require.NoError(t, err, "context must have account ID set via EnsureAccount")

	install := &app.Install{
		ID:          domains.NewInstallID(),
		Name:        generics.GetFakeObj[string](),
		AppID:       appID,
		OrgID:       org.ID,
		CreatedByID: accountID,
	}

	res := s.db.WithContext(ctx).Create(install)
	require.NoError(t, res.Error)

	return install
}
