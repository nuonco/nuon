package job

// this is a workflow that is used to execute a job. It is designed to be reusable outside the context of this
// namespace, and for all jobs. Thus, it has it's own activities, and other components to allow it to work more
// effectively.
type jobWorkflow struct{}
