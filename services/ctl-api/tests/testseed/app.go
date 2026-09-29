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

func BuildComponent(appID string) *app.Component {
	id := domains.NewComponentID()
	acct := BuildAccount()
	return &app.Component{
		ID:                id,
		Name:              fmt.Sprintf("component_%s", id),
		AppID:             appID,
		CreatedByID:       acct.ID,
		Status:            "queued",
		StatusDescription: "waiting for queue to provision component",
	}
}

func (s *Seeder) CreateComponent(ctx context.Context, t *testing.T, appID string, componentType app.ComponentType) *app.Component {
	c := &app.Component{
		Name:              fmt.Sprintf("component_%s", domains.NewComponentID()),
		AppID:             appID,
		Type:              componentType,
		Status:            "queued",
		StatusDescription: "waiting for queue to provision component",
	}
	res := s.db.WithContext(ctx).Create(c)
	require.NoError(t, res.Error)
	return c
}
