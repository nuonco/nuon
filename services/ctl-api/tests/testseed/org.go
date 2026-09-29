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

func BuildOrg() *app.Org {
	acct := BuildAccount()
	id := domains.NewOrgID()
	return &app.Org{
		ID:          id,
		Name:        fmt.Sprintf("org-%s", id),
		OrgType:     app.OrgTypeSandbox,
		Status:      app.OrgStatusActive,
		SandboxMode: true,
		CreatedBy:   *acct,
		CreatedByID: acct.ID,
	}
}

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

func (s *Seeder) EnsureOrg(ctx context.Context, t *testing.T) (context.Context, *app.Org) {
	org := s.CreateOrg(ctx, t)
	return cctx.SetOrgIDContext(ctx, org.ID), org
}
