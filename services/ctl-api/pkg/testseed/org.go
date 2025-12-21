package testseed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// EnsureOrg creates a test organization and sets it in the context.
// Returns the updated context with the org ID set.
//
// The organization is created with:
// - A unique fake name
// - OrgType: Sandbox
// - Status: Active
// - SandboxMode: true (no real cloud resources)
//
// The returned context has the org ID set via cctx.SetOrgIDContext,
// which is required for most ctl-api operations that are scoped to an org.
func (s *Seeder) EnsureOrg(ctx context.Context, t *testing.T) context.Context {
	org := app.Org{
		Name:        generics.GetFakeObj[string](),
		OrgType:     app.OrgTypeSandbox,
		Status:      app.OrgStatusActive,
		SandboxMode: true,
	}
	res := s.db.WithContext(ctx).Create(&org)
	require.Nil(t, res.Error)

	return cctx.SetOrgIDContext(ctx, org.ID)
}
