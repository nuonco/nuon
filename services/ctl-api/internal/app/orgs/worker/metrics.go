package worker

import (
	"strconv"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) defaultTags(orgID string, sandboxMode bool) map[string]string {
	return map[string]string{
		"namespace":    "orgs",
		"sandbox_mode": strconv.FormatBool(sandboxMode),
		"org_id":       orgID,
	}
}

func (w *Workflows) writeStatusMetric(ctx workflow.Context, name string, err error, tags map[string]string, addtlTags ...string) {
	tags["status"] = "ok"
	if err != nil {
		tags["status"] = "error"
	}

	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.metricsWriter.Incr(name, 1, metrics.ToTags(tags, addtlTags...))
		return nil
	})
}

func (w *Workflows) writeIncrMetric(ctx workflow.Context, name string, tags map[string]string, addtlTags ...string) {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.metricsWriter.Incr(name, 1, metrics.ToTags(tags, addtlTags...))
		return nil
	})
}
