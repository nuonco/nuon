package check

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if h.state.cfg.Noop {
		l.Info("Noop")
		return nil
	}

	response := map[string]time.Duration{}
	l.Info("started healthcheck")
	l.Info(fmt.Sprintf("checking on %d job loops", len(h.jobLoops)))
	for _, jl := range h.jobLoops {
		hc, jobGroup := jl.GetHealthcheck()
		l.Info(fmt.Sprintf("%+v", hc), zap.String("group", jobGroup))
		response[string(jobGroup)] = jl.TimeSinceLastHealthcheck()
		jl.SetLatestHealthcheckAt()
	}
	h.state.outputs = configs.HealthcheckOutputs{JobLoops: response}
	l.Info(fmt.Sprintf("%+v", response))

	l.Info("completed healthcheck")

	return nil
}
