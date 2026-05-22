package emitter

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

type EmitterWorkflowRequest struct {
	QueueID   string `validate:"required"`
	EmitterID string `validate:"required"`
	Version   string `validate:"required"`

	State *EmitterState
}

// EmitterState is the data that is passed between continue-as-news
type EmitterState struct {
	EmitCount int64
}

// @temporal-gen-v2 workflow
// @task-queue "queue"
// @id-template queue-emitter-{{.QueueID}}-{{.EmitterID}}
func (w *Workflows) Emitter(ctx workflow.Context, req EmitterWorkflowRequest) error {
	e := &emitterWorkflow{
		cfg:       w.cfg,
		v:         w.v,
		db:        w.db,
		tClient:   w.tClient,
		l:         w.l,
		mw:        w.mw,
		emitterID: req.EmitterID,
		queueID:   req.QueueID,
		state:     req.State,
	}
	if e.state == nil {
		e.state = &EmitterState{
			EmitCount: 0,
		}
	}

	finished, err := e.run(ctx)
	if err != nil {
		return err
	}
	if !finished {
		req.State = e.state
		return workflow.NewContinueAsNewError(ctx, w.Emitter, req)
	}

	// For cron-scheduled emitters, returning nil just completes this run —
	// Temporal will schedule the next one. We must terminate the workflow
	// to actually stop the cron.
	info := workflow.GetInfo(ctx)
	_ = activities.AwaitTerminateWorkflow(ctx, &activities.TerminateWorkflowRequest{
		WorkflowID: info.WorkflowExecution.ID,
		Namespace:  info.Namespace,
		Reason:     "emitter finished",
	})

	return nil
}

type emitterWorkflow struct {
	cfg     *internal.Config
	v       *validator.Validate
	db      *gorm.DB
	tClient temporalclient.Client
	l       *zap.Logger
	mw      tmetrics.Writer

	emitterID string
	queueID   string

	stopped   bool
	restarted bool

	state *EmitterState
}
