package worker

import (
	"github.com/powertoolsdev/mono/pkg/config"
)

// each domain has it's own namespace, so we no longer need to split work by domain at the task queue level. By having a
// single queue in each namespace, we can more easily understand queue depth + have headroom to have a
// high-priority/low-priority queue in the future.
const DefaultTaskQueue string = "main"

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_task_queue", DefaultTaskQueue)
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_max_concurrent_activities", 1)
}

// Config defines the standard workflow worker config, which all workers should embed as part of their application.
type Config struct {
	// builtin configuration
	Env         config.Env `config:"env" validate:"required"`
	ServiceName string     `config:"service_name" validate:"required"`

	// temporal configuration
	TemporalHost                    string `config:"temporal_host" validate:"required"`
	TemporalNamespace               string `config:"temporal_namespace" validate:"required"`
	TemporalTaskQueue               string `config:"temporal_task_queue" validate:"required"`
	TemporalMaxConcurrentActivities int    `config:"temporal_max_concurrent_activities" validate:"required"`

	// observability configuration
	HostIP   string `config:"host_ip" validate:"required"`
	LogLevel string `config:"log_level"`
}
