package workflows

import "github.com/powertoolsdev/mono/services/ctl-api/internal/app"

type WorkflowStepOptions func(*app.WorkflowStep)

func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}
