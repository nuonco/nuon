package drainvm

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
	pkgshutdown "github.com/nuonco/nuon/bins/runner/internal/pkg/shutdown"
)

const pollTimeout = 65 * time.Minute

// Exec drains the runner process by sending SIGUSR1, then waits for it to exit.
func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// Prevent the monitor from restarting nuon-runner.service after it exits
	h.monitor.SetDraining()

	pid, err := getRunnerPID()
	if err != nil {
		l.Warn("unable to get runner PID, skipping drain", zap.Error(err))
		return nil
	}

	if pid == 0 {
		l.Info("runner process not running, skipping drain")
		return nil
	}

	l.Info("sending SIGUSR1 to runner process", zap.Int("pid", pid))
	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		l.Warn("unable to send SIGUSR1", zap.Error(err))
		return nil
	}

	l.Info("waiting for runner process to exit", zap.Duration("timeout", pollTimeout))
	if err := waitForServiceInactive(ctx, l, pollTimeout); err != nil {
		l.Warn("runner process did not exit in time, proceeding with shutdown", zap.Error(err))
	} else {
		l.Info("runner process exited cleanly")
	}

	return nil
}

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
		Status: models.AppRunnerJobStatusFinished,
	}); err != nil {
		return err
	}

	return pkgshutdown.Shutdown(ctx, l, h.v)
}

func getRunnerPID() (int, error) {
	out, err := exec.Command(
		"systemctl", "show", "-p", "MainPID", "--value",
		monitor.RunnerServiceName,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("systemctl show: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parse PID: %w", err)
	}
	return pid, nil
}

func waitForServiceInactive(ctx context.Context, l *zap.Logger, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timed out waiting for %s to become inactive", monitor.RunnerServiceName)
		case <-ticker.C:
			out, err := exec.Command("systemctl", "is-active", monitor.RunnerServiceName).Output()
			if err != nil {
				// is-active returns non-zero for inactive, which is what we want
				status := strings.TrimSpace(string(out))
				if status == "inactive" || status == "failed" || status == "dead" {
					return nil
				}
			}
			status := strings.TrimSpace(string(out))
			l.Debug("runner service still active", zap.String("status", status))
		}
	}
}
