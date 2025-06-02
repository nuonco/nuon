package app

// TODO(sdboyer) remove
type InstallWorkflowContext struct {
	ID             string
	WorkflowStepID *string
}

type FlowContext struct {
	ID             string
	FlowStepID *string // TODO(sdboyer) rename this to just StepID when things are working well enough that we can tell what breaks
	StepName *string
}
