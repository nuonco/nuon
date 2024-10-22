package propagator

import (
	"context"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

var _ workflow.ContextPropagator = (*optionalPropagator)(nil)

type optionalPropagator struct {
	l          *zap.Logger
	propagator workflow.ContextPropagator
}

func (s *optionalPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	if err := s.propagator.Inject(ctx, writer); err != nil {
		s.l.Debug("propagator inject failed", zap.Error(err))
	}

	return nil
}

func (s *optionalPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	if err := s.propagator.InjectFromWorkflow(ctx, writer); err != nil {
		s.l.Debug("propagator inject from workflow failed", zap.Error(err))
	}

	return nil
}

func (s *optionalPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	newCtx, err := s.propagator.Extract(ctx, reader)
	if err != nil {
		s.l.Debug("propagator extract failed", zap.Error(err))
		return ctx, nil
	}

	return newCtx, nil
}

func (s *optionalPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	newCtx, err := s.propagator.ExtractToWorkflow(ctx, reader)
	if err != nil {
		s.l.Debug("propagator extract to workflow failed", zap.Error(err))
		return ctx, nil
	}

	return newCtx, nil
}
