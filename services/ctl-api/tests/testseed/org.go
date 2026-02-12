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

// BuildOrg creates an app.Org with fake defaults.
func BuildOrg() *app.Org {
	acct := BuildAccount()
	return &app.Org{
		ID:          domains.NewOrgID(),
		Name:        generics.GetFakeObj[string](),
		OrgType:     app.OrgTypeSandbox,
		Status:      app.OrgStatusActive,
		SandboxMode: true,
		CreatedBy:   *acct,
		CreatedByID: acct.ID,
	}
}

// CreateOrg builds and persists an org to the database.
// Uses account from context if available for CreatedByID.
func (s *Seeder) CreateOrg(ctx context.Context, t *testing.T) *app.Org {
	org := BuildOrg()
	if accountID, err := cctx.AccountIDFromContext(ctx); err == nil {
		org.CreatedBy = app.Account{}
		org.CreatedByID = accountID
	}
	res := s.db.WithContext(ctx).Create(org)
	require.NoError(t, res.Error)
	return org
}

// EnsureOrg creates a test org and sets its ID in the context via cctx.SetOrgIDContext.
func (s *Seeder) EnsureOrg(ctx context.Context, t *testing.T) (context.Context, *app.Org) {
	org := s.CreateOrg(ctx, t)
	return cctx.SetOrgIDContext(ctx, org.ID), org
}
