package componenthealth

import "context"

// ReportOnce runs a single report cycle and returns any error from it,
// without starting the engine's background ticker. It is for callers that
// supply their own schedule -- e.g. the control-plane install component
// health workflow, which drives one report per Temporal activity invocation
// instead of a long-lived process loop. The Runner's own New()/run() path is
// unchanged by this addition.
func (e *Engine) ReportOnce(ctx context.Context) error {
	reportCtx, cancel := context.WithTimeout(ctx, reportTimeout)
	defer cancel()
	return e.report(reportCtx)
}
