package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/taskqueue"
)

func seedOwnerQueues(t testing.TB, db *gorm.DB, ownerType, namespace, ownerID, orgID, createdByID string) {
	t.Helper()
	specs, ok := queuenames.Specs(ownerType)
	require.True(t, ok)
	for _, spec := range specs {
		q := &app.Queue{
			CreatedByID: createdByID,
			OrgID:       &orgID,
			OwnerID:     ownerID,
			OwnerType:   ownerType,
			Name:        spec.Name,
			MaxInFlight: spec.MaxInFlight,
			MaxDepth:    spec.MaxDepth,
			Workflow: signaldb.WorkflowRef{
				Namespace:  namespace,
				IDTemplate: "queue-%s",
				TaskQueue:  taskqueue.For(namespace, spec.Name),
			},
		}
		require.NoError(t, db.Clauses(clause.OnConflict{DoNothing: true}).Create(q).Error)
	}
}

func seedInstallQueues(t testing.TB, db *gorm.DB, install *app.Install) {
	t.Helper()
	seedOwnerQueues(t, db, queuenames.OwnerInstalls, "installs", install.ID, install.OrgID, install.CreatedByID)
}

func seedOrgQueues(t testing.TB, db *gorm.DB, org *app.Org) {
	t.Helper()
	seedOwnerQueues(t, db, queuenames.OwnerOrgs, "orgs", org.ID, org.ID, org.CreatedByID)
}
