package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type orgServiceMockWorkflowRun struct{}

func (*orgServiceMockWorkflowRun) GetID() string                          { return "mock-workflow-id" }
func (*orgServiceMockWorkflowRun) GetRunID() string                       { return "mock-run-id" }
func (*orgServiceMockWorkflowRun) Get(context.Context, interface{}) error { return nil }
func (*orgServiceMockWorkflowRun) GetWithOptions(context.Context, interface{}, tclient.WorkflowRunGetOptions) error {
	return nil
}

func resetOrgServiceSignals(t *testing.T, db *gorm.DB) {
	require.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&app.QueueSignal{}).Error)
}

func closeOrgServiceDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func seedOrgServiceFixtures(t *testing.T, db *gorm.DB, account *app.Account, org *app.Org) {
	ctx := cctx.SetAccountContext(context.Background(), account)
	orgID := org.ID

	var role app.Role
	result := db.Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeOrgAdmin}).First(&role)
	if result.Error == gorm.ErrRecordNotFound {
		role = app.Role{
			OrgID:    generics.NewNullString(orgID),
			RoleType: app.RoleTypeOrgAdmin,
			Contexts: []string{app.RoleContextTeam},
			Managed:  true,
		}
		require.NoError(t, db.WithContext(ctx).Create(&role).Error)
	} else {
		require.NoError(t, result.Error)
		if !role.AllowsContext(app.RoleContextTeam) {
			role.Contexts = append(role.Contexts, app.RoleContextTeam)
			require.NoError(t, db.WithContext(ctx).Save(&role).Error)
		}
	}

	queue := app.Queue{
		OrgID:       &orgID,
		OwnerID:     orgID,
		OwnerType:   "orgs",
		Name:        orgshelpers.OrgSignalsQueueName,
		MaxDepth:    50,
		MaxInFlight: 10,
	}
	require.NoError(t, db.WithContext(ctx).Where(app.Queue{
		OwnerID:   queue.OwnerID,
		OwnerType: queue.OwnerType,
		Name:      queue.Name,
	}).FirstOrCreate(&queue).Error)
}
