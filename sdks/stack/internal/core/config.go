// Package core holds the method-agnostic types shared between the public
// stack package and the individual provisioning method implementations
// (internal/awssdk, internal/terraform, internal/cloudformation). It carries
// no cloud-provider SDK dependencies so that every method package — and the
// public API — can import it without creating an import cycle.
package core

// Config carries the per-install rendered configuration that the ctl-api
// produces alongside a stack run. Fields common to every cloud live here;
// cloud-specific inputs live on a per-cloud sub-struct (AWS, GCP, …), exactly
// one of which is populated, selected by Cloud. Each cloud's sub-struct
// mirrors the corresponding install-stacks module's variables.tf.
type Config struct {
	// Cloud selects the target cloud provider. Empty falls back to
	// DefaultCloud (aws), so configs produced before the field existed keep
	// their AWS behavior.
	Cloud Cloud `json:"cloud,omitempty"`

	// InstallID is duplicated here so config can be threaded through the
	// provisioning methods without dragging method-specific state along.
	InstallID string `json:"install_id,omitempty"`
	OrgID     string `json:"org_id,omitempty"`
	AppID     string `json:"app_id,omitempty"`

	RunnerID     string `json:"runner_id,omitempty"`
	RunnerAPIURL string `json:"runner_api_url,omitempty"`

	// PhoneHomeURL is the endpoint the Terraform module's phone-home reports
	// to. The Terraform method renders it into tfvars and lets the module
	// report the run; the AWS SDK method ignores it (the SDK reports directly).
	PhoneHomeURL string `json:"phone_home_url,omitempty"`

	// Method selects the provisioning implementation (sdk or terraform).
	// Empty falls back to the cloud's default (see DefaultMethodForCloud).
	Method Method `json:"method,omitempty"`

	// Terraform* configure the terraform method; ignored by other methods.
	// TerraformVersion empty resolves the latest stable release at runtime.
	// TerraformModuleURL empty defaults to the install-stacks main archive,
	// TerraformModuleSubdir empty defaults to the cloud's module subdir.
	TerraformVersion      string `json:"terraform_version,omitempty"`
	TerraformModuleURL    string `json:"terraform_module_url,omitempty"`
	TerraformModuleSubdir string `json:"terraform_module_subdir,omitempty"`

	// InstallInputs, AutoGenerateSecrets and Secrets are cloud-agnostic in
	// both shape and semantics: every cloud's module consumes the same
	// install_inputs map and the same auto_generate_secrets / secrets inputs.
	InstallInputs       map[string]string      `json:"install_inputs,omitempty"`
	AutoGenerateSecrets []string               `json:"auto_generate_secrets,omitempty"`
	Secrets             map[string]SecretInput `json:"secrets,omitempty"`

	// RequiredInputs lists the names of install inputs that must have a
	// non-empty value before provisioning. InstallInputs is a plain map and
	// carries no per-key metadata, so the required set is tracked separately;
	// the SDK enforces it at provision time.
	RequiredInputs []string `json:"required_inputs,omitempty"`

	// AWS carries the AWS-specific inputs; populated when Cloud is aws.
	AWS *AWSConfig `json:"aws,omitempty"`
	// GCP carries the GCP-specific inputs; populated when Cloud is gcp.
	GCP *GCPConfig `json:"gcp,omitempty"`
}

// SecretInput mirrors the customer-provided secret shape. Identical across
// clouds, so it lives in the common config rather than a per-cloud file.
type SecretInput struct {
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Value       string `json:"value,omitempty"`
}

// Prefix is the resource-name prefix used across the stack. Matches the TF
// module's `local.prefix = var.nuon_install_id` exactly. Why we don't add
// "nuon-": ctl-api and downstream app templates derive role / log-group /
// secret names from the install id directly (e.g. the runner's IID
// validation expects `{install_id}-runner` as the role name); double-
// prefixing breaks every cross-system lookup.
func (c *Config) Prefix() string {
	return c.InstallID
}
