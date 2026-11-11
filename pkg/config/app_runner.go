package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"

	"github.com/nuonco/nuon/pkg/config/source"
)

type AppRunnerConfig struct {
	Source string `mapstructure:"source,omitempty" toml:"source,omitempty"`

	RunnerType string `mapstructure:"runner_type" toml:"runner_type" jsonschema:"required"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`

	HelmDriver string `mapstructure:"helm_driver,omitempty" toml:"helm_driver,omitempty"`

	InitScriptURL string `mapstructure:"init_script_url" toml:"init_script_url"`

	// PhoneHomeScriptURL overrides the phone-home Lambda source for this app.
	PhoneHomeScriptURL string `mapstructure:"phone_home_script_url,omitempty" toml:"phone_home_script_url,omitempty"`

	// InstanceType sets the cloud machine/instance type for the install runner host.
	InstanceType string `mapstructure:"instance_type,omitempty" toml:"instance_type,omitempty"`

	// ImageAutoUpdate resolves the latest runner image at every instance launch instead of once
	// at stack create/update. Ignored when ImageID is set.
	ImageAutoUpdate bool `mapstructure:"image_auto_update,omitempty" toml:"image_auto_update,omitempty"`

	// ImageID pins the install runner host to a specific image (an AMI ID for aws).
	ImageID string `mapstructure:"image_id,omitempty" toml:"image_id,omitempty"`

	// RunnerAPIURL overrides the Nuon runner API endpoint for installs using this config.
	RunnerAPIURL string `mapstructure:"runner_api_url,omitempty" toml:"runner_api_url,omitempty"`

	// PublicAPIURL overrides the Nuon public API endpoint used for phone-home callbacks.
	PublicAPIURL string `mapstructure:"public_api_url,omitempty" toml:"public_api_url,omitempty"`

	// Deprecated
	EnvVars []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var"`
}

func (a AppRunnerConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("external configuration source").
		Long("Path to an external file containing runner configuration (YAML, JSON, or TOML)").
		Field("runner_type").Short("type of runner").Required().
		Long("Specifies how the runner executes deployments").
		Example("aws").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables to pass to the runner as key-value pairs").
		Example("DEBUG=true").
		Example("LOG_LEVEL=info").
		Field("helm_driver").Short("Helm driver configuration").
		Long("Specifies the backend driver for Helm operations (e.g., 'configmap', 'secret')").
		Example("configmap").
		Example("secret").
		Field("init_script_url").Short("initialization script URL").
		Long("URL to a script that runs during runner initialization. Supports HTTP(S), git, file, and relative paths (./). Examples: https://example.com/script.sh, ./scripts/init.sh, git::https://github.com/org/repo//script.sh, file:///path/to/script.sh").
		Example("https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init-mng-v2.sh").
		Field("phone_home_script_url").Short("phone-home Lambda source URL").
		Long("URL to the Python source for the install stack's phone-home Lambda, fetched when the stack template is rendered and embedded inline. Leave unset to use the pinned default. Set this to move a single app onto a different script version without affecting other apps.").
		Example("https://raw.githubusercontent.com/nuonco/runner/refs/tags/aws-v0.1.4/scripts/aws/phonehome.py").
		Field("instance_type").Short("machine/instance type for the install runner").
		Long("Cloud machine/instance type used for the install runner host. Cloud-specific value mapped per runner_type (an EC2 instance type for aws, a machine type for gcp, a VM size for azure). Defaults to the platform default when unset").
		Example("t3a.medium").
		Example("e2-medium").
		Example("Standard_D2s_v5").
		Field("image_auto_update").Short("resolve the latest runner image at every launch").
		Long("When true, the install runner host picks up the latest base image every time an instance launches. When false (the default), the image is resolved once when the install stack is created or updated. Ignored when image_id is set. Only applies to aws runners whose stack runner template supports it; otherwise it is ignored").
		Example("true").
		Field("image_id").Short("pin the runner host image").
		Long("Pins the install runner host to a specific image, overriding image_auto_update. Use it to hold runners on a known-good image. For aws this is an AMI ID. Only applies to aws runners whose stack runner template supports it; otherwise it is ignored").
		Example("ami-0abcdef1234567890").
		Field("runner_api_url").Short("custom runner API endpoint").
		Long("Set this to proxy all runner API calls (heartbeats, job polling, logs) through your own domain instead of Nuon's.").
		Example("https://runner-api.example.com").
		Field("public_api_url").Short("custom public API endpoint").
		Long("Set this to proxy phone-home calls through your own domain instead of Nuon's.").
		Example("https://runner-api.example.com").
		Field("env_var").Short("deprecated: use env_vars map instead").
		Long("Deprecated: Array of name/value pairs for environment variables. Use the env_vars map instead")
}

func (a *AppRunnerConfig) parse(cfg ParseOptions) error {
	if a == nil {
		return ErrConfig{
			Description: "an app runner config is required",
		}
	}

	if len(a.EnvVars) > 0 {
		return ErrConfig{
			Description: "env_var arrays are deprecated, please use env_vars map instead",
		}
	}

	if a.Source != "" {
		obj, err := source.LoadSourceFrom(a.Source, cfg.RootDir, cfg.Source)
		if err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("unable to load source %s", a.Source),
				Err:         err,
			}
		}

		if err := mapstructure.Decode(obj, &a); err != nil {
			return err
		}
	}

	return a.validateImageID()
}

var awsAMIIDPattern = regexp.MustCompile(`^ami-[0-9a-f]+$`)

// Only aws image IDs are checked: other clouds use their own formats and don't consume image_id yet.
func (a *AppRunnerConfig) validateImageID() error {
	if !strings.HasPrefix(a.RunnerType, "aws") {
		return nil
	}
	return ValidateAWSImageID("image_id", a.ImageID)
}

// ValidateAWSImageID checks that a non-empty runner image ID is an AMI ID.
func ValidateAWSImageID(field, imageID string) error {
	if imageID == "" || awsAMIIDPattern.MatchString(imageID) {
		return nil
	}
	return ErrConfig{
		Description: fmt.Sprintf("%s %q is not a valid AMI ID (expected ami- followed by hex characters)", field, imageID),
	}
}
