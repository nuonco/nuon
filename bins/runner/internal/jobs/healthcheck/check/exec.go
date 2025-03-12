package check

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

// health check: iterate through all the job loops and check their stats, set a value, and log the response
// if a job loop is not running, we should restart it
func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// NOTE: if this is a noop healthcheck, return immediately
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
