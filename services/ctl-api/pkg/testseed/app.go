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

func (s *Seeder) EnsureApp(ctx context.Context, t *testing.T) *app.App {
	org, err := cctx.OrgFromContext(ctx)
	require.NoError(t, err, "context must have org set via EnsureOrg")

	accountID, err := cctx.AccountIDFromContext(ctx)
	require.NoError(t, err, "context must have account ID set via EnsureAccount")

	testApp := &app.App{
		ID:          domains.NewAppID(),
		Name:        generics.GetFakeObj[string](),
		OrgID:       org.ID,
		CreatedByID: accountID,
	}

	res := s.db.WithContext(ctx).Create(testApp)
	require.NoError(t, res.Error)

	return testApp
}
