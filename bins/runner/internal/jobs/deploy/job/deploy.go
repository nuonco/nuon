package job

import (
	"context"

	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (p *handler) deploy(
	ctx context.Context,
) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	clientset, err := p.getClientset()
	if err != nil {
		return err
	}

	l.Info("starting job")
	job, err := p.startJob(ctx, clientset)
	if err != nil {
		return err
	}

	l.Info("polling job")
	err = p.pollJob(ctx, clientset, job)
	if err != nil {
		l.Error("error polling job: %v", zap.Error(err))
		return err
	}

	return nil
}
