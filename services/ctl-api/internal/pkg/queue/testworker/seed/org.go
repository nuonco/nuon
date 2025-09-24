package seed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

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
