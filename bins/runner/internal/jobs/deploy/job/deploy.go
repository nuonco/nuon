package job

import (
	"context"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	jobv1 "github.com/powertoolsdev/mono/pkg/types/plugins/job/v1"
	"go.uber.org/zap"
)

func (p *handler) deploy(
	ctx context.Context,
) (*jobv1.Deployment, error) {
	// init logger
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	clientset, err := p.getClientset()
	if err != nil {
		return nil, err
	}

	// start k8s job
	l.Info("starting job")
	job, err := p.startJob(ctx, clientset)
	if err != nil {
		return nil, err
	}

	// monitor job
	l.Info("polling job")
	err = p.pollJob(ctx, clientset, job)
	if err != nil {
		l.Error("error polling job: %v", zap.Error(err))
		return nil, err
	}

	return &jobv1.Deployment{}, nil
}
