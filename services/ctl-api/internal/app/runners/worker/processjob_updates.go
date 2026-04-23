package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/processjobupdates"
)

// processJobState holds the flags toggled by ProcessJob's update handlers.
// Both startJobExecution and monitorJobExecution share one instance so the
// ProcessJob function can register handlers once at entry and let the
// helpers workflow.Await on the flags.
//
// Flags are cleared by the helper reading them, not by the handler.
type processJobState struct {
	jobStatusChanged    bool
	execCreated         bool
	execStatusChanged   bool
	runnerStatusChanged bool
	runnerRestarted     bool
}

func (s *processJobState) anyFlag() bool {
	return s.jobStatusChanged ||
		s.execCreated ||
		s.execStatusChanged ||
		s.runnerStatusChanged ||
		s.runnerRestarted
}

func (s *processJobState) clear() {
	s.jobStatusChanged = false
	s.execCreated = false
	s.execStatusChanged = false
	s.runnerStatusChanged = false
	s.runnerRestarted = false
}

// registerProcessJobUpdateHandlers wires the five push-update handlers onto
// the caller's workflow context. Must be called at the top of ProcessJob so
// external writers can deliver updates before either helper begins waiting.
func registerProcessJobUpdateHandlers(ctx workflow.Context, s *processJobState) error {
	if err := workflow.SetUpdateHandlerWithOptions(ctx, processjobupdates.UpdateNameJobStatusChanged,
		func(ctx workflow.Context, _ processjobupdates.JobStatusChangedPayload) error {
			s.jobStatusChanged = true
			return nil
		}, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, processjobupdates.UpdateNameJobExecutionCreated,
		func(ctx workflow.Context, _ processjobupdates.JobExecutionCreatedPayload) error {
			s.execCreated = true
			return nil
		}, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, processjobupdates.UpdateNameJobExecutionStatus,
		func(ctx workflow.Context, _ processjobupdates.JobExecutionStatusPayload) error {
			s.execStatusChanged = true
			return nil
		}, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, processjobupdates.UpdateNameRunnerStatusChanged,
		func(ctx workflow.Context, _ processjobupdates.RunnerStatusChangedPayload) error {
			s.runnerStatusChanged = true
			return nil
		}, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	return workflow.SetUpdateHandlerWithOptions(ctx, processjobupdates.UpdateNameRunnerRestarted,
		func(ctx workflow.Context, _ processjobupdates.RunnerRestartedPayload) error {
			s.runnerRestarted = true
			return nil
		}, workflow.UpdateHandlerOptions{})
}
