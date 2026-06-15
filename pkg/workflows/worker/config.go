package worker

import (
	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/pkg/workflows"
)

const (
	defaultMaxConcurrentActivities                int = 10000
	defaultMaxConcurrentWorkflowTaskExecutionSize int = 10000
	defaultMaxConcurrentActivityTaskPollers       int = 20
	defaultMaxConcurrentWorkflowTaskPollers       int = 20
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_task_queue", workflows.DefaultTaskQueue)
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_max_concurrent_activities", defaultMaxConcurrentActivities)
	config.RegisterDefault("temporal_max_concurrent_workflow_task_execution_size", defaultMaxConcurrentWorkflowTaskExecutionSize)
	config.RegisterDefault("temporal_max_concurrent_activity_task_pollers", defaultMaxConcurrentActivityTaskPollers)
	config.RegisterDefault("temporal_max_concurrent_workflow_task_pollers", defaultMaxConcurrentWorkflowTaskPollers)
}

// Config defines the standard workflow worker config, which all workers should embed as part of their application.
type Config struct {
	// builtin configuration
	Env         config.Env `config:"env" validate:"required"`
	ServiceName string     `config:"service_name" validate:"required"`

	GitRef  string `config:"git_ref" validate:"required"`
	Version string `config:"version" validate:"required"`

	// temporal configuration
	TemporalHost                                   string `config:"temporal_host" validate:"required"`
	TemporalNamespace                              string `config:"temporal_namespace"`
	TemporalNamespacePrefix                        string `config:"temporal_namespace_prefix"`
	TemporalTLSEnabled                             bool   `config:"temporal_tls_enabled"`
	TemporalAPIKey                                 string `config:"temporal_api_key"`
	TemporalTaskQueue                              string `config:"temporal_task_queue" validate:"required"`
	TemporalMaxConcurrentActivities                int    `config:"temporal_max_concurrent_activities" validate:"required" faker:"oneof: 10,20"`
	TemporalMaxConcurrentWorkflowTaskExecutionSize int    `config:"temporal_max_concurrent_workflow_task_execution_size" validate:"required" faker:"oneof: 10,20"`
	TemporalMaxConcurrentActivityTaskPollers       int    `config:"temporal_max_concurrent_activity_task_pollers" validate:"required" faker:"oneof: 10,20"`
	TemporalMaxConcurrentWorkflowTaskPollers       int    `config:"temporal_max_concurrent_workflow_task_pollers" validate:"required" faker:"oneof: 10,20"`

	// observability configuration
	HostIP               string `config:"host_ip" validate:"required"`
	LogLevel             string `config:"log_level"`
	SlowQueryThresholdMS int    `config:"slow_query_threshold_ms"`
}

// NamespaceFor resolves a short logical namespace (e.g. "installs", "orgs")
// to the fully-qualified Temporal namespace. Empty TemporalNamespacePrefix
// returns the short name unchanged (in-cluster Temporal). When set (e.g.
// "<install-name>"), returns "<prefix>-<short>", matching the nuon-lite
// Temporal Cloud naming scheme created by the temporal-namespaces component.
func (c Config) NamespaceFor(short string) string {
	if c.TemporalNamespacePrefix == "" {
		return short
	}
	return c.TemporalNamespacePrefix + "-" + short
}
