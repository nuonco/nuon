package runner

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) execAWSActivity(ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	if err := w.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, resp); err != nil {
		return err
	}

	return nil
}
