package check

import "context"

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	// NOTE: this specificially only "serializes" the values we want to store in the db
	jobloops := make(map[string]interface{})
	// Copy values from h.state.outputs to result
	for k, v := range h.state.outputs.JobLoops {
		jobloops[k] = v
	}
	outputs := map[string]interface{}{"job_loops": jobloops}
	return outputs, nil
}
