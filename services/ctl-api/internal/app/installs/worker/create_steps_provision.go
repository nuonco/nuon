package worker

import "go.temporal.io/sdk/workflow"

func (w *Workflows) createStepsProvision(ctx workflow.Context) error {
	// need to fetch the install app-config-id

	// create install cloudformation stack
	// await cloudformation stack finished
	// await runner online

	// provision sandbox

	// for each component
	// create pre-component-deploy-hook
	// create deploy
	// create post-component-deploy-hook

	return nil
}
