package worker

import (
	"github.com/powertoolsdev/mono/pkg/config"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
)

const (
	defaultMaxConcurrentActivities int = 10
)

//nolint:gochecknoinits
func init() {
	config.RegisterDefault("temporal_task_queue", workflowsclient.DefaultTaskQueue)
	config.RegisterDefault("temporal_host", "localhost:7233")
	config.RegisterDefault("temporal_max_concurrent_activities", defaultMaxConcurrentActivities)
}

// Config defines the standard workflow worker config, which all workers should embed as part of their application.
type Config struct {
	// builtin configuration
	Env         config.Env `config:"env" validate:"required"`
	ServiceName string     `config:"service_name" validate:"required"`

	GitRef  string `config:"git_ref" validate:"required"`
	Version string `config:"version" validate:"required"`

	// temporal configuration
	TemporalHost                    string `config:"temporal_host" validate:"required"`
	TemporalNamespace               string `config:"temporal_namespace" validate:"required"`
	TemporalTaskQueue               string `config:"temporal_task_queue" validate:"required"`
	TemporalMaxConcurrentActivities int    `config:"temporal_max_concurrent_activities" validate:"required" faker:"oneof: 10,20"`

	// observability configuration
	HostIP   string `config:"host_ip" validate:"required"`
	LogLevel string `config:"log_level"`
}
