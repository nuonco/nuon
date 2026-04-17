package seed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker/example"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type EnsureWorkflowRequest struct {
	OwnerID   string
	OwnerType string
	Type      app.WorkflowType
	Steps     []example.StepConfig
	PlanOnly  bool
}

type EnsureWorkflowResult struct {
	Workflow *app.Workflow
}

func (s *Seeder) EnsureWorkflow(ctx context.Context, t *testing.T, req EnsureWorkflowRequest) *EnsureWorkflowResult {
	wfType := req.Type
	if wfType == "" {
		wfType = app.WorkflowTypeManualDeploy
	}

	flw := app.Workflow{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Type:      wfType,
		Status:    app.CompositeStatus{Status: app.StatusPending},
		PlanOnly:  req.PlanOnly,
		GenerateStepsSignal: &signaldb.SignalData{
			Signal: &example.FakeGenerateStepsSignal{
				Steps: req.Steps,
			},
		},
	}

	res := s.db.WithContext(ctx).Create(&flw)
	require.Nil(t, res.Error)

	return &EnsureWorkflowResult{
		Workflow: &flw,
	}
}
