package flowutil

// DirectiveKey is the metadata key used to communicate step execution directives
// from the step workflow back to the parent conductor.
const DirectiveKey = "directive"

// Step directives tell the parent conductor how to proceed after a step completes.
const (
	DirectiveContinue      = "continue"
	DirectiveSkipGroup     = "skip-group"
	DirectiveStop          = "stop"
	DirectiveRetry         = "retry"
	DirectiveAwaitApproval = "await-approval"
)
