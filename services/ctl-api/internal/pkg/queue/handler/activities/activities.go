package activities

import (
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

type Params struct {
	fx.In

	V       *validator.Validate
	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
}

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB `name:"psql"`
	tclient temporalclient.Client
}

func New(params Params) *Activities {
	return &Activities{
		v:       params.V,
		db:      params.DB,
		tclient: params.TClient,
	}
}

// handlerStartOperation builds a WithStartWorkflowOperation for a handler workflow.
// This ensures the handler workflow is running when we send an update to it.
// If the handler was terminated, this will start a new one; if it's already running,
// it will use the existing one.
func (a *Activities) handlerStartOperation(workflowID string, queueID string, queueSignalID string) tclient.WithStartWorkflowOperation {
	req := handler.HandlerRequest{
		QueueID:       queueID,
		QueueSignalID: queueSignalID,
	}
	startOpts := tclient.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                "api",
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	return a.tclient.NewWithStartWorkflowOperation(startOpts, "Handler", req)
}
