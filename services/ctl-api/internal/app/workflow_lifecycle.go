package app

import "github.com/nuonco/nuon/pkg/lifecyclephase"

func (w *Workflow) InstallLifecyclePhase() *lifecyclephase.LifecyclePhase {
	failed := w.Status.Status == StatusError || w.Status.Status == StatusCancelled

	switch w.Type {
	case WorkflowTypeProvision:
		phase := lifecyclephase.Provisioned
		description := "Provision workflow completed"
		if failed {
			description = "Provision workflow failed"
		}
		// A stack-only provision deliberately leaves the sandbox and components
		// unprovisioned, so it must not read as fully provisioned.
		if w.IsStackOnly() && !failed {
			phase = lifecyclephase.Provisioning
			description = "Stack and runner ready, waiting to provision the sandbox and components"
		}
		return &lifecyclephase.LifecyclePhase{Phase: phase, Description: description}
	case WorkflowTypeReprovision:
		if w.PlanOnly {
			return nil
		}
		description := "Reprovision workflow completed"
		if failed {
			description = "Reprovision workflow failed"
		}
		return &lifecyclephase.LifecyclePhase{Phase: lifecyclephase.Provisioned, Description: description}
	case WorkflowTypeDeprovision:
		description := "Deprovision workflow completed"
		if failed {
			description = "Deprovision workflow failed"
		}
		return &lifecyclephase.LifecyclePhase{Phase: lifecyclephase.Deprovisioned, Description: description}
	}
	return nil
}
