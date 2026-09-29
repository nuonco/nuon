package check

import "context"

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	jobloops := make(map[string]interface{})
	for k, v := range h.state.outputs.JobLoops {
		jobloops[k] = v
	}
	outputs := map[string]interface{}{"job_loops": jobloops}
	return outputs, nil
}
