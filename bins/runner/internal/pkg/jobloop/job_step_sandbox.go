package jobloop

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	logPeriod  time.Duration = time.Second / 4
	totalSteps               = 6
)

func (j *jobLoop) execSandboxStep(ctx context.Context, job *models.AppRunnerJob) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if state := j.sandboxCtl.GetState(); state != nil {
		if state.CheckAndClearPanic() {
			l.Error("sandbox control: panic requested")
			panic("sandbox: panic requested via sandbox control API")
		}
		if state.CheckAndClearShutdown() {
			l.Error("sandbox control: shutdown requested")
			j.shutdowner.Shutdown()
			return errors.New("sandbox: shutdown requested via sandbox control API")
		}
	}

	jobType := string(job.Type)
	duration := j.cfg.SandboxJobDuration
	shouldFault := rand.Intn(10) == 0
	faultsEnabled := j.cfg.SandboxModeFaultsEnabled
	var faultMessage string

	if state := j.sandboxCtl.GetState(); state != nil {
		cfg := state.GetConfig(jobType)
		duration = cfg.Duration
		if cfg.FaultRate > 0 {
			shouldFault = rand.Float64() < cfg.FaultRate
			faultsEnabled = true
		} else if cfg.FaultRate == 0 && cfg.Preset != "default" {
			shouldFault = false
		}
		if cfg.ErrorMessage != "" {
			faultMessage = cfg.ErrorMessage
		}
	}

	stepDuration := duration / totalSteps
	l.Info("sandbox mode enabled, faking job output",
		zap.String("step", "initialize"),
		zap.Duration("duration", duration),
		zap.String("job_type", jobType),
	)

	if shouldFault {
		l.Error("sandbox mode fault selected, will return an error at the end of this job")
	}

	timeout := time.NewTimer(stepDuration)
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

	if shouldFault && faultsEnabled {
		if state := j.sandboxCtl.GetState(); state != nil {
			state.RecordResult(false)
		}
		if faultMessage != "" {
			return errors.New(faultMessage)
		}
		return errors.New("Sandbox Mode Fault Injected")
	}

	if state := j.sandboxCtl.GetState(); state != nil {
		state.RecordResult(true)
	}

	return nil
}
