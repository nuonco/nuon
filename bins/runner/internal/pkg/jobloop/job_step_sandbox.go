package jobloop

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	logPeriod  time.Duration = time.Second / 4
	totalSteps               = 6
)

func (j *jobLoop) execSandboxStep(ctx context.Context) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	duration := j.cfg.SandboxJobDuration / totalSteps
	l.Info("sandbox mode enabled, faking job output",
		zap.String("step", "initialize"),
		zap.Duration("duration", j.cfg.SandboxJobDuration),
	)

	shouldFault := rand.Intn(10) == 0
	l.Error("sandbox mode fault randomly selected, will return an error at the end of this job")

	timeout := time.NewTimer(duration)
	ticker := time.NewTicker(logPeriod)
	defer ticker.Stop()
	defer timeout.Stop()

	for {
		select {
		case <-ticker.C:
			l.Info("sandbox job log",
				zap.String("key", "value"),
				zap.Any("obj", map[string]interface{}{}),
			)
		case <-timeout.C:
			goto BREAK
		}
	}
BREAK:
	l.Info("sandbox job log ending",
		zap.String("key", "value"),
		zap.Any("obj", map[string]interface{}{}),
	)

	if shouldFault && j.cfg.SandboxModeFaultsEnabled {
		return errors.New("Sandbox Mode Fault Injected")
	}

	return nil
}
