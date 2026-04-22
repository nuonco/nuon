package service

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// loadStepLogStreams loads log streams for workflow steps that have step_target_type "install_workflow_steps".
// These steps (e.g., sync-secrets) own their log stream directly rather than through a separate target entity.
func (s *service) loadStepLogStreams(ctx context.Context, steps []*app.WorkflowStep) {
	var stepIDs []string
	stepMap := make(map[string]*app.WorkflowStep)
	for _, step := range steps {
		if step.StepTargetType == "install_workflow_steps" && step.StepTargetID != "" {
			stepIDs = append(stepIDs, step.StepTargetID)
			stepMap[step.StepTargetID] = step
		}
	}

	if len(stepIDs) == 0 {
		return
	}

	var logStreams []app.LogStream
	s.db.WithContext(ctx).
		Where(app.LogStream{OwnerType: "install_workflow_steps"}).
		Where("owner_id IN ?", stepIDs).
		Order("created_at DESC").
		Find(&logStreams)

	for i := range logStreams {
		ls := &logStreams[i]
		if step, ok := stepMap[ls.OwnerID]; ok {
			step.LogStream = ls
		}
	}
}
