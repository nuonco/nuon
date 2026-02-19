package testseed

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// BuildApp creates an app.App with fake defaults.
func BuildApp() *app.App {
	org := BuildOrg()
	acct := BuildAccount()

	id := domains.NewAppID()

	return &app.App{
		ID:          id,
		Name:        fmt.Sprintf("app-%s", id),
		Org:         org,
		OrgID:       org.ID,
		CreatedBy:   *acct,
		CreatedByID: acct.ID,
	}
}

// CreateApp builds and persists an app to the database.
// Uses org/account from context if available.
func (s *Seeder) CreateApp(ctx context.Context, t *testing.T) *app.App {
	a := BuildApp()
	if orgID, err := cctx.OrgIDFromContext(ctx); err == nil {
		a.Org = nil
		a.OrgID = orgID
	}
	if accountID, err := cctx.AccountIDFromContext(ctx); err == nil {
		a.CreatedBy = app.Account{}
		a.CreatedByID = accountID
	}
	res := s.db.WithContext(ctx).Create(a)
	require.NoError(t, res.Error)
	return a
}
