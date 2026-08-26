// Package core holds the method-agnostic types shared between the public
// stack package and the individual provisioning method implementations
// (internal/terraform). It carries
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
	// to. The module renders it into tfvars and reports the run.
	PhoneHomeURL string `json:"phone_home_url,omitempty"`

	// Method is retained for wire compatibility; Terraform is currently the
	// only provisioning implementation. Empty falls back to the default.
	Method Method `json:"method,omitempty"`

	// TerraformVersion empty resolves the latest stable release at runtime.
	// TerraformModuleURL empty defaults to the install-stacks main archive,
	// TerraformModuleSubdir empty defaults to the cloud's module subdir.
	TerraformVersion      string `json:"terraform_version,omitempty"`
	TerraformModuleURL    string `json:"terraform_module_url,omitempty"`
	TerraformModuleSubdir string `json:"terraform_module_subdir,omitempty"`

	// TerraformExecPath, when set, is an existing terraform binary to run
	// instead of downloading one via hc-install. Lets locked-down or airgapped
	// callers (e.g. the Terraform provider) avoid the releases.hashicorp.com
	// fetch.
	TerraformExecPath string `json:"terraform_exec_path,omitempty"`

	// TerraformWorkDir, when set, is the directory the terraform method assembles
	// the module + tfvars in and runs terraform from. Empty means a fresh
	// per-run temp dir. State lives in the remote backend, so the work dir is
	// disposable.
	TerraformWorkDir string `json:"terraform_work_dir,omitempty"`

	// TerraformBackend, when set, configures a remote state backend for the
	// terraform method (S3 for AWS, GCS for GCP). Empty leaves terraform on its
	// default local state.
	TerraformBackend *TerraformBackend `json:"terraform_backend,omitempty"`

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

// TerraformBackend configures the remote state backend for the terraform
// method. It is stored in the customer's target account: S3 for AWS, GCS for
// GCP, keyed per install so state survives across applies and ephemeral hosts.
// The customer supplies an existing bucket.
type TerraformBackend struct {
	// Bucket is the S3 (AWS) or GCS (GCP) bucket holding state. Required.
	Bucket string `json:"bucket,omitempty"`

	// Key is the S3 object key (AWS only). Empty defaults to
	// nuon/<install_id>/terraform.tfstate.
	Key string `json:"key,omitempty"`
	// Region is the S3 bucket region (AWS only). Empty falls back to the
	// install's AWS region.
	Region string `json:"region,omitempty"`
	// DynamoDBTable is an optional S3 state-lock table (AWS only). Empty relies
	// on S3 native locking.
	DynamoDBTable string `json:"dynamodb_table,omitempty"`

	// Prefix is the GCS object prefix (GCP only). Empty defaults to
	// nuon/<install_id>.
	Prefix string `json:"prefix,omitempty"`
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
