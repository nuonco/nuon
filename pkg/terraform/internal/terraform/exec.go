package terraform

import "context"

// Init executes `init`
func (w *workspace) Init(ctx context.Context) error {
	return w.tfExecutor.Init(ctx)
}

// Plan executes `plan`
func (w *workspace) Plan(ctx context.Context) error {
	return w.tfExecutor.Plan(ctx)
}

// Apply executes `apply`
func (w *workspace) Apply(ctx context.Context) error {
	return w.tfExecutor.Apply(ctx)
}

// Destroy executes `destroy`
func (w *workspace) Destroy(ctx context.Context) error {
	return w.tfExecutor.Destroy(ctx)
}

// Output executes `output`
func (w *workspace) Output(ctx context.Context) (map[string]interface{}, error) {
	return w.tfExecutor.Output(ctx)
}
