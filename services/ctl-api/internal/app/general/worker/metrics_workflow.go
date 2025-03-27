package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"

	enumsv1 "go.temporal.io/api/enums/v1"
)

const (
	metricsWorkflowCronTab string = "*/1 * * * *"
	metricsWorkflowName    string = "general-metrics-cron"
)

func (w *Workflows) startMetricsWorkflow(ctx workflow.Context) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            metricsWorkflowName,
		CronSchedule:          metricsWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.Metrics)
}

func (w *Workflows) Metrics(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("deadman.snitch", zap.String("status", "running"))

	methods := map[string]func(workflow.Context) error{
		"psql_tables": func(ctx workflow.Context) error {
			return w.writePSQLTableMetrics(ctx)
		},
		"clickhouse_tables": func(ctx workflow.Context) error {
			return w.writeCHTableMetrics(ctx)
		},
		"temporal_orgs": func(ctx workflow.Context) error {
			return w.temporalNamespaceMetrics(ctx, "orgs")
		},
		"temporal_apps": func(ctx workflow.Context) error {
			return w.temporalNamespaceMetrics(ctx, "apps")
		},
		"temporal_components": func(ctx workflow.Context) error {
			return w.temporalNamespaceMetrics(ctx, "components")
		},
		"temporal_installs": func(ctx workflow.Context) error {
			return w.temporalNamespaceMetrics(ctx, "installs")
		},
		"temporal_releases": func(ctx workflow.Context) error {
			return w.temporalNamespaceMetrics(ctx, "releases")
		},
		"temporal_runners": func(ctx workflow.Context) error {
			return w.temporalNamespaceMetrics(ctx, "runners")
		},
	}

	for name, method := range methods {
		if err := method(ctx); err != nil {
			l.Error("error executing metrics step", zap.String("name", name))
			return errors.Wrap(err, "unable to execute step "+name)
		}
	}

	return nil
}

func (w *Workflows) writeCHTableMetrics(ctx workflow.Context) error {
	defaultTags := map[string]string{"general": "true"}

	// write psql tables
	tables, err := activities.AwaitGetCHTableMetrics(ctx, activities.GetCHTableMetricsRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to get table metrics")
	}
	for _, table := range tables {
		w.mw.Gauge(ctx, "table_size", table.SizeBytes, metrics.ToTags(generics.MergeMap(map[string]string{
			"db_type":    "ch",
			"table_name": table.TableName,
		}, defaultTags))...)
	}

	return nil
}

func (w *Workflows) writePSQLTableMetrics(ctx workflow.Context) error {
	defaultTags := map[string]string{"general": "true"}

	// write psql tables
	tables, err := activities.AwaitGetPSQLTableMetrics(ctx, activities.GetPSQLTableMetricsRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to get table metrics")
	}
	for _, table := range tables {
		w.mw.Gauge(ctx, "table_size", table.SizeBytes, metrics.ToTags(generics.MergeMap(map[string]string{
			"db_type":    "psql",
			"table_name": table.TableName,
		}, defaultTags))...)
	}

	return nil
}

func (w *Workflows) temporalNamespaceMetrics(ctx workflow.Context, ns string) error {
	defaultTags := map[string]string{"general": "true", "namespace": ns}

	m, err := activities.AwaitGetNamespaceMetricsByName(ctx, ns)
	if err != nil {
		return errors.Wrap(err, "unable to get metrics")
	}

	w.mw.Gauge(ctx, "eventloops.count",
		float64(m.EventLoops),
		metrics.ToTags(generics.MergeMap(map[string]string{}, defaultTags))...)

	w.mw.Gauge(ctx, "workflows.count",
		float64(m.AllWorkflows),
		metrics.ToTags(generics.MergeMap(map[string]string{
			"workflow_type": "event_loop",
		}, defaultTags))...)

	w.mw.Gauge(ctx, "eventloops.expected_count",
		float64(m.ExpectedEventLoops),
		metrics.ToTags(generics.MergeMap(map[string]string{}, defaultTags))...)

	return nil
}
