package job

import (
	"context"

	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (p *handler) deploy(
	ctx context.Context,
) error {
	// init logger
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	clientset, err := p.getClientset()
	if err != nil {
		return err
	}

	// start k8s job
	l.Info("starting job")
	job, err := p.startJob(ctx, clientset)
	if err != nil {
		return err
	}

	// monitor job
	l.Info("polling job")
	err = p.pollJob(ctx, clientset, job)
	if err != nil {
		l.Error("error polling job: %v", zap.Error(err))
		return err
	}

	return nil
}
