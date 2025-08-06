package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/flow"
)

// @temporal-gen workflow
// @execution-timeout 720h
func (w *Workflows) RerunFlow(ctx workflow.Context, sreq signals.RequestSignal) error {
	if sreq.FlowID == "" {
		sreq.FlowID = sreq.InstallWorkflowID
	}
	fc := &flow.WorkflowConductor[*signals.Signal]{
		Cfg:        w.cfg,
		V:          w.v,
		MW:         w.mw,
		Generators: w.getWorkflowStepGenerators(ctx),
		ExecFn:     w.getExecuteFlowExecFn(sreq),
	}

	return fc.Rerun(ctx, sreq.EventLoopRequest, flow.RerunInput{
		FlowID:       sreq.FlowID,
		StepID:       sreq.RerunConfiguration.StepID,
		StalePlan:    sreq.RerunConfiguration.StalePlan,
		RePlanStepID: sreq.RerunConfiguration.RePlanStepID,
		Operation:    flow.RerunOperation(sreq.RerunConfiguration.StepOperation),
	})
}
