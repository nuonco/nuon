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

// EnsureInstall creates a test install in the database.
// Returns a pointer to the created install.
//
// The install is created with:
// - A unique fake name
// - A unique fake ID (26-character)
// - Associated with the given app
// - Associated with the org from context
// - Created by the account from context
//
// The context must already have both org and account set via EnsureOrg and EnsureAccount.
func (s *Seeder) EnsureInstall(ctx context.Context, t *testing.T, appID string) *app.Install {
	// Extract org ID from context
	org, err := cctx.OrgFromContext(ctx)
	require.NoError(t, err, "context must have org set via EnsureOrg")

	// Extract account ID from context
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
