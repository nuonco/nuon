package terraform

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
)

func (p *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Tag this handler's logger with semantic-convention attributes so every
	// emitted record (including from helpers further down the call tree) carries
	// them automatically.
	l = l.With(
		zap.String("service.name", "runner.sandbox.sync_secrets"),
		zap.String("nuon.tool", "sync_secrets"),
		zap.String("nuon.deploy.kind", "sandbox.sync_secrets"),
		zap.String("sync_secrets.operation", string(job.Operation)),
	)
	ctx = pkgctx.SetLogger(ctx, l)

	for _, secret := range p.state.plan.KubernetesSecrets {
		l.Info("syncing secret " + secret.Name)
		opCtx, end := op.Tool(ctx, "sync_secrets", "sync")
		err := p.execSyncSecret(opCtx, secret)
		end(err)
		if err != nil {
			return errors.Wrap(err, "unable to sync secret")
		}
	}

	return nil
}
