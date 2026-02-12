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

// BuildInstall creates an app.Install with fake defaults.
func BuildInstall() *app.Install {
	org := BuildOrg()
	acct := BuildAccount()
	builtApp := BuildApp()
	builtApp.Org = org
	builtApp.OrgID = org.ID
	builtApp.CreatedByID = acct.ID

	return &app.Install{
		ID:          domains.NewInstallID(),
		Name:        generics.GetFakeObj[string](),
		Org:         *org,
		OrgID:       org.ID,
		CreatedBy:   *acct,
		CreatedByID: acct.ID,
		App:         *builtApp,
		AppID:       builtApp.ID,
	}
}

// CreateInstall builds and persists an install to the database.
// Uses org/account from context if available.
func (s *Seeder) CreateInstall(ctx context.Context, t *testing.T) *app.Install {
	i := BuildInstall()
	if orgID, err := cctx.OrgIDFromContext(ctx); err == nil {
		i.Org = app.Org{}
		i.OrgID = orgID
	}
	if accountID, err := cctx.AccountIDFromContext(ctx); err == nil {
		i.CreatedBy = app.Account{}
		i.CreatedByID = accountID
	}
	res := s.db.WithContext(ctx).Create(i)
	require.NoError(t, res.Error)
	return i
}
