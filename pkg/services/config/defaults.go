package config

// Env is an environment string
type Env string

const (
	// Production is the production environment
	Production Env = "production"
	// Development is a development environment not local
	Development Env = "development"
	// Stage is a staging environment
	Stage Env = "stage"

	defaultPort          = 9102
	defaultSampleRate    = 1.0
	defaultMaxBatchCount = 256
)

// UnmarshalConfig unmarshals a config value string to the associated interface
// type
func (e *Env) UnmarshalConfig(value string) {
	switch value {
	case "development", "dev":
		*e = Development
	case "production", "prod":
		*e = Production
	case "stage", "staging":
		*e = Stage
	default:
		*e = Development
	}
}

func (e *Env) String() string {
	return string(*e)
}

// Version is set by ldflags, do not set / use this directly
// use the value exposed from the Base config
var Version string = "unknown"

// init registers the defaults
//
//nolint:gochecknoinits
func init() {
	RegisterDefault("system_port", defaultPort)
	RegisterDefault("export_runtime_metrics", true)
	RegisterDefault("trace_sample_rate", defaultSampleRate)
	RegisterDefault("trace_max_batch_count", defaultMaxBatchCount)
	RegisterDefault("version", Version)
}

// Base is the base configuration for all services
type Base struct {
	Env                          Env     `config:"env"`
	LogLevel                     string  `config:"log_level"`
	ExportRuntimeMetrics         bool    `config:"export_runtime_metrics"`
	ProjectID                    string  `config:"project_id"`
	ServiceOwner                 string  `config:"service_owner"`
	ServiceName                  string  `config:"service_name"`
	SystemPort                   int     `config:"system_port"`
	TraceAddress                 string  `config:"host_ip"`
	TraceMaxBatchCount           int     `config:"trace_max_batch_count"`
	TraceSampleRate              float64 `config:"trace_sample_rate"`
	ProfilerEnabled              bool    `config:"profiler_enabled"`
	ProfilerBlockProfileRate     int     `config:"profiler_block_profile_rate"`
	ProfilerMutexProfileFraction int     `config:"profiler_mutex_profile_fraction"`
	DisableLogSampling           bool    `config:"disable_log_sampling"`
	DisableStackTraces           bool    `config:"disable_stack_traces"`
	Version                      string  `config:"version"`
}
