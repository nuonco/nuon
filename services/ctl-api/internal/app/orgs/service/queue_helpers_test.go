package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"

	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func createOrgSignalsQueue(t testing.TB, db *gorm.DB, acc *app.Account, orgID string) *app.Queue {
	ctx := cctx.SetAccountContext(context.Background(), acc)
	q := &app.Queue{
		OwnerID:     orgID,
		OwnerType:   "orgs",
		Name:        orgshelpers.OrgSignalsQueueName,
		MaxDepth:    50,
		MaxInFlight: 10,
	}
	q.Workflow.Namespace = "orgs"
	q.Workflow.IDTemplate = "queue-%s"
	require.NoError(t, db.WithContext(ctx).Create(q).Error)
	return q
}

func createOrgRoles(t testing.TB, client *authz.Client, acc *app.Account, orgID string) {
	ctx := cctx.SetAccountContext(context.Background(), acc)
	require.NoError(t, client.CreateOrgRoles(ctx, orgID))
}

type mockWorkflowRun struct{}

func (m *mockWorkflowRun) GetID() string                                { return "mock-workflow-id" }
func (m *mockWorkflowRun) GetRunID() string                             { return "mock-run-id" }
func (m *mockWorkflowRun) Get(ctx context.Context, v interface{}) error { return nil }
func (m *mockWorkflowRun) GetWithOptions(ctx context.Context, v interface{}, o tclient.WorkflowRunGetOptions) error {
	return nil
}

func newPermissiveTemporalMock(t *testing.T) *temporal.MockClient {
	mockTC := temporal.NewMockClient(gomock.NewController(t))
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&mockWorkflowRun{}, nil).AnyTimes()
	return mockTC
}
