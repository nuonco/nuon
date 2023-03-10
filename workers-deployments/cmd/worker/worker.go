package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	shared "github.com/powertoolsdev/workers-deployments/internal"
)

type worker struct {
	v *validator.Validate `validate:"required"`

	// TODO(jm): once we have a standard configuration library, accept that
	Config *shared.Config `validate:"required"`

	Workflows  []interface{} `validate:"required"`
	Activities []interface{} `validate:"required"`
}

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

func WithConfig(cfg *shared.Config) workerOption {
	return func(w *worker) error {
		w.Config = cfg
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
	w.registerWorker(wkr)

	if err := wkr.Run(interruptCh); err != nil {
		return fmt.Errorf("error running worker: %w", err)
	}
	return nil
}
