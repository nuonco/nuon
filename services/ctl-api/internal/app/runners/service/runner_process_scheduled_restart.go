package service

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
)

func (s *service) processUptimeThreshold(processType app.RunnerProcessType) (time.Duration, bool) {
	switch processType {
	case app.RunnerProcessTypeInstall:
		if s.cfg.ProcessInstallUptimeThreshold > 0 {
			return s.cfg.ProcessInstallUptimeThreshold, true
		}
		return helpers.DefaultInstallUptimeThreshold, true
	case app.RunnerProcessTypeMng:
		if s.cfg.ProcessMngUptimeThreshold > 0 {
			return s.cfg.ProcessMngUptimeThreshold, true
		}
		return helpers.DefaultMngUptimeThreshold, true
	default:
		return 0, false
	}
}

func (s *service) scheduledRestartAt(p *app.RunnerProcess) *time.Time {
	if p == nil || p.StartedAt == nil {
		return nil
	}
	threshold, ok := s.processUptimeThreshold(p.Type)
	if !ok {
		return nil
	}
	t := p.StartedAt.Add(threshold)
	return &t
}

func (s *service) attachScheduledRestarts(processes []*app.RunnerProcess) {
	for _, p := range processes {
		if t := s.scheduledRestartAt(p); t != nil {
			p.NextScheduledRestartAt = t
		}
	}
}

func runnerProcessPtrs(processes []app.RunnerProcess) []*app.RunnerProcess {
	ptrs := make([]*app.RunnerProcess, len(processes))
	for i := range processes {
		ptrs[i] = &processes[i]
	}
	return ptrs
}
