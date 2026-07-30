package config

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

const (
	HealthProbeTypeHTTP = "http"
	HealthProbeTypeTCP  = "tcp"
	HealthProbeTypeExec = "exec"
)

// ComponentHealthConfig configures live health checking for component types the
// health engine can observe (helm_chart and kubernetes_manifest).
type ComponentHealthConfig struct {
	Enabled             *bool                        `mapstructure:"enabled,omitempty" toml:"enabled,omitempty" nuonhash:"omitempty"`
	StabilizationWindow string                       `mapstructure:"stabilization_window,omitempty" toml:"stabilization_window,omitempty" features:"template" nuonhash:"omitempty"`
	BlockDeploy         *bool                        `mapstructure:"block_deploy,omitempty" toml:"block_deploy,omitempty" nuonhash:"omitempty"`
	Probes              []ComponentHealthProbeConfig `mapstructure:"probes,omitempty" toml:"probes,omitempty" nuonhash:"omitempty"`
}

// ComponentHealthProbeConfig is one synthetic health check the runner executes
// for the component on every health report cycle.
type ComponentHealthProbeConfig struct {
	Type    string   `mapstructure:"type,omitempty" toml:"type,omitempty" nuonhash:"omitempty"`
	Name    string   `mapstructure:"name,omitempty" toml:"name,omitempty" nuonhash:"omitempty"`
	URL     string   `mapstructure:"url,omitempty" toml:"url,omitempty" features:"template" nuonhash:"omitempty"`
	Command []string `mapstructure:"command,omitempty" toml:"command,omitempty" features:"template" nuonhash:"omitempty"`
}

func (c ComponentHealthConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("enabled").Short("enable live health checking").
		Long("Whether live health checking and the health verdict apply to this component. Default: true").
		Default("true").
		Example("false").
		Field("stabilization_window").Short("health stabilization window").
		Long("How long the component must hold healthy after a deploy applies before the deploy step is considered done. Duration string (e.g., \"3m\", \"10m\"). Default: 3m. Max: 1h").
		Default("3m").
		Example("3m").
		Example("10m").
		Field("block_deploy").Short("fail the deploy when health does not stabilize").
		Long("When true, the deploy step fails if health does not stabilize inside the window. When false, the step still completes and only records what health did. Default: false").
		Default("false").
		Example("true").
		Field("probes").Short("synthetic health probes").
		Long("Probes the runner executes from inside the install to assert the component is actually serving. Each probe reports as its own health resource, and a failing probe makes the component unhealthy")
}

func (p ComponentHealthProbeConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("probe type").
		Long("How the probe checks the component. http issues a GET and treats any 2xx/3xx as healthy, tcp opens a connection, exec runs a command inside the runner and treats exit code 0 as healthy").
		Enum(HealthProbeTypeHTTP, HealthProbeTypeTCP, HealthProbeTypeExec).
		Example("http").
		Example("tcp").
		Example("exec").
		Field("name").Short("probe display name").
		Long("Name shown for this probe in the dashboard and CLI. Defaults to the url, or the first element of command for exec probes").
		Example("api-healthz").
		Field("url").Short("probe target").
		Long("Target for http and tcp probes. http requires a full URL, tcp accepts host:port or a URL whose scheme implies the port. Supports Nuon templating").
		Example("https://{{.nuon.install.sandbox.outputs.public_domain}}/healthz").
		Example("db.internal:5432").
		Field("command").Short("exec probe command").
		Long("Command to run for an exec probe, as an argv array. Executed directly with no shell, so pipes and redirection are not interpreted. Exit code 0 is healthy. Supports Nuon templating").
		Example([]string{"/usr/local/bin/check-api", "--timeout", "2s"})
}

// Validate checks the probes declared on the health block. Anything it rejects
// would otherwise reach the runner as a probe it cannot execute.
func (c *ComponentHealthConfig) Validate() error {
	if c == nil {
		return nil
	}

	for idx := range c.Probes {
		if err := c.Probes[idx].Validate(); err != nil {
			return fmt.Errorf("health.probes[%d]: %w", idx, err)
		}
	}

	return nil
}

func (p ComponentHealthProbeConfig) Validate() error {
	switch strings.ToLower(strings.TrimSpace(p.Type)) {
	case HealthProbeTypeHTTP, HealthProbeTypeTCP:
		if strings.TrimSpace(p.URL) == "" {
			return fmt.Errorf("%s probes require a url", p.Type)
		}
	case HealthProbeTypeExec:
		if len(p.Command) == 0 || strings.TrimSpace(p.Command[0]) == "" {
			return fmt.Errorf("exec probes require a command")
		}
	case "":
		return fmt.Errorf("type is required, and must be one of %s, %s, %s",
			HealthProbeTypeHTTP, HealthProbeTypeTCP, HealthProbeTypeExec)
	default:
		return fmt.Errorf("invalid type %q, must be one of %s, %s, %s",
			p.Type, HealthProbeTypeHTTP, HealthProbeTypeTCP, HealthProbeTypeExec)
	}

	return nil
}
