package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/flow"
)

// @temporal-gen workflow
// @execution-timeout 720h
// @id-template {{.Req.ID}}-execute-workflow-{{.Req.InstallWorkflowID}}-rerun-flow
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

	err := fc.Rerun(ctx, sreq.EventLoopRequest, flow.RerunInput{
		ContinueFromIdx: sreq.StartFromStepIdx,
		FlowID:          sreq.FlowID,
		StepID:          sreq.RerunConfiguration.StepID,
		StalePlan:       sreq.RerunConfiguration.StalePlan,
		RePlanStepID:    sreq.RerunConfiguration.RePlanStepID,
		Operation:       flow.RerunOperation(sreq.RerunConfiguration.StepOperation),
	})
	if err != nil {
		cerr, ok := err.(*flow.ContinueAsNewErr)
		if ok && cerr != nil {
			sreq.StartFromStepIdx = cerr.StartFromStepIdx
			return workflow.NewContinueAsNewError(ctx, w.RerunFlow, sreq)
		}
		return err
	}

	return nil
}
