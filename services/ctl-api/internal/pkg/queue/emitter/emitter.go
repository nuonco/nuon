package emitter

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type EmitterWorkflowRequest struct {
	EmitterID string
	Version   string

	State *EmitterState
}

// EmitterState is the data that is passed between continue-as-news
type EmitterState struct {
	EmitCount int64
}

// @temporal-gen workflow
// @task-queue "queue"
// @id-template queue-emitter-{{.EmitterID}}
func (w *Workflows) Emitter(ctx workflow.Context, req EmitterWorkflowRequest) error {
	e := &emitterWorkflow{
		cfg:       w.cfg,
		v:         w.v,
		db:        w.db,
		tClient:   w.tClient,
		l:         w.l,
		emitterID: req.EmitterID,
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

	return nil
}

type emitterWorkflow struct {
	cfg     *internal.Config
	v       *validator.Validate
	db      *gorm.DB
	tClient temporalclient.Client
	l       *zap.Logger

	emitterID string

	stopped   bool
	restarted bool

	state *EmitterState
}
