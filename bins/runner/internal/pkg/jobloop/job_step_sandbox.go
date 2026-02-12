package jobloop

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/sandboxctl"
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

	// Check for failure mode from sandboxctl
	if j.sandboxCtl != nil && j.sandboxCtl.Active() {
		state := j.sandboxCtl.State()
		switch state.GetFailureMode() {
		case sandboxctl.FailureModeError:
			time.Sleep(500 * time.Millisecond)
			l.Error("sandbox control: failure mode is error, returning error")
			return errors.New("sandbox control: simulated error failure")
		case sandboxctl.FailureModePanic:
			time.Sleep(500 * time.Millisecond)
			l.Error("sandbox control: failure mode is panic, panicking")
			panic("sandbox control: simulated panic failure")
		case sandboxctl.FailureModeShutdown:
			time.Sleep(500 * time.Millisecond)
			l.Error("sandbox control: failure mode is shutdown, shutting down")
			j.shutdowner.Shutdown()
			return errors.New("sandbox control: simulated shutdown failure")
		}
	}

	// Determine job duration: sandboxctl override > config default
	jobDuration := j.cfg.SandboxJobDuration
	if j.sandboxCtl != nil && j.sandboxCtl.Active() {
		if override := j.sandboxCtl.State().GetJobDuration(); override > 0 {
			jobDuration = override
		}
	}

	duration := jobDuration / totalSteps
	l.Info("sandbox mode enabled, faking job output",
		zap.String("step", "initialize"),
		zap.Duration("duration", jobDuration),
	)

	// Determine faults: sandboxctl override > config default
	faultsEnabled := j.cfg.SandboxModeFaultsEnabled
	if j.sandboxCtl != nil && j.sandboxCtl.Active() {
		faultsEnabled = j.sandboxCtl.State().GetFaultsEnabled()
	}

	shouldFault := rand.Intn(10) == 0
	if shouldFault && faultsEnabled {
		l.Error("sandbox mode fault randomly selected, will return an error at the end of this job")
	}

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

	if shouldFault && faultsEnabled {
		return errors.New("Sandbox Mode Fault Injected")
	}

	return nil
}
