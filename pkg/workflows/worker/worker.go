package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

// Mode defines what does worker the run
type Mode string

const (
	// ModeAll runs both workflows and activities
	ModeAll Mode = "all"
	// ModeWorkflows runs only workflows
	ModeWorkflows Mode = "workflows"
	// ModeActivities runs only activities
	ModeActivities Mode = "activities"
)

type worker struct {
	v  *validator.Validate `validate:"required"`
	mw metrics.Writer
	l  *zap.Logger

	Config      *Config       `validate:"required"`
	Workflows   []interface{} `validate:"required,gt=0"`
	Activities  []interface{} `validate:"required,gt=0"`
	Namespace   string        `validate:"required"`
	propagators []workflow.ContextPropagator
	Mode        Mode
}

type Worker interface {
	Run(<-chan interface{}) error
}

var _ Worker = (*worker)(nil)

func New(v *validator.Validate, opts ...workerOption) (*worker, error) {
	wkr := &worker{
		v:          v,
		Workflows:  make([]interface{}, 0),
		Activities: make([]interface{}, 0),
	}
	for _, opt := range opts {
		if err := opt(wkr); err != nil {
			return nil, err
		}
	}

	if err := wkr.v.Struct(wkr); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	if err := wkr.v.Struct(wkr.Config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return wkr, nil
}

type workerOption func(*worker) error

func WithConfig(cfg *Config) workerOption {
	return func(w *worker) error {
		w.Config = cfg
		w.Namespace = cfg.TemporalNamespace
		return nil
	}
}

func WithNamespace(ns string) workerOption {
	return func(w *worker) error {
		w.Namespace = ns
		return nil
	}
}

func WithWorkflow(wkflow interface{}) workerOption {
	return func(w *worker) error {
		w.Workflows = append(w.Workflows, wkflow)
		return nil
	}
}

func WithActivity(act interface{}) workerOption {
	return func(w *worker) error {
		w.Activities = append(w.Activities, act)
		return nil
	}
}

func WithContextPropagator(propagator workflow.ContextPropagator) workerOption {
	return func(t *worker) error {
		t.propagators = append(t.propagators, propagator)
		return nil
	}
}

func WithLogger(l *zap.Logger) workerOption {
	return func(w *worker) error {
		w.l = l
		return nil
	}
}

func WithMetricsWriter(mw metrics.Writer) workerOption {
	return func(w *worker) error {
		w.mw = mw
		return nil
	}
}

func WithMode(mode string) workerOption {
	return func(w *worker) error {
		w.Mode = Mode(mode)
		return nil
	}
}

func (w *worker) Run(interruptCh <-chan interface{}) error {
	client, closeFn, err := w.getClient()
	if err != nil {
		return fmt.Errorf("unable to get client: %w", err)
	}
	defer closeFn()

	wkr, err := w.getWorker(client)
	if err != nil {
		return fmt.Errorf("unable to get worker: %w", err)
	}

	switch w.Mode {
	case ModeAll:
		w.registerActivitiesOnWorker(wkr)
		w.registerWorkflowsOnWorker(wkr)
	case ModeWorkflows:
		w.registerWorkflowsOnWorker(wkr)
	case ModeActivities:
		w.registerActivitiesOnWorker(wkr)
	}

	if err := wkr.Run(interruptCh); err != nil {
		return fmt.Errorf("error running worker: %w", err)
	}
	return nil
}
