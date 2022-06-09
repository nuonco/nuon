package config

// Env is an environment string
type Env string

const (
	// Local is a local development environment
	Local Env = "local"
	// Demo is the demo environment
	Demo Env = "demo"
	// QA is the QA environment
	QA Env = "qa"
	// Production is the production environment
	Production Env = "production"
	// Development is a development environment not local
	Development Env = "development"
	// Staging is a staging environment
	Staging Env = "staging"
)

// UnmarshalConfig unmarshals a config value string to the associated interface
// type
func (e *Env) UnmarshalConfig(value string) {
	switch value {
	case "demo":
		*e = Demo
	case "development", "dev":
		*e = Development
	case "qa":
		*e = QA
	case "production", "prod":
		*e = Production
	case "stage", "staging":
		*e = Staging
	default:
		*e = Local
	}
}

func (e Env) String() string {
	switch e {
	case Demo:
		return "demo"
	case Development:
		return "development"
	case Staging:
		return "stage"
	case QA:
		return "qa"
	default:
		return "local"
	}
}

// init registers the defaults
func init() {
	RegisterDefault("env", Local)
	RegisterDefault("system_port", 9102)
	RegisterDefault("export_runtime_metrics", true)
	RegisterDefault("trace_sample_rate", 1.0)
	RegisterDefault("trace_max_batch_count", 256)
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
	TraceAddress                 string  `config:"trace_address"`
	TraceMaxBatchCount           int     `config:"trace_max_batch_count"`
	TraceSampleRate              float64 `config:"trace_sample_rate"`
	ProfilerEnabled              bool    `config:"profiler_enabled"`
	ProfilerBlockProfileRate     int     `config:"profiler_block_profile_rate"`
	ProfilerMutexProfileFraction int     `config:"profiler_mutex_profile_fraction"`
	DisableLogSampling           bool    `config:"disable_log_sampling"`
	DisableStackTraces           bool    `config:"disable_stack_traces"`
}
