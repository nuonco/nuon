package terraform

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	for _, secret := range p.state.plan.KubernetesSecrets {
		l.Info("syncing secret " + secret.Name)
		if err := p.execSyncSecret(ctx, secret); err != nil {
			return errors.Wrap(err, "unable to sync secret")
		}
	}

	return nil
}
