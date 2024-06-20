package configs

import (
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

type TerraformDeployRunType string

const (
	TerraformDeployRunTypePlan    TerraformDeployRunType = "plan"
	TerraformDeployRunTypeDestroy TerraformDeployRunType = "destroy"
	TerraformDeployRunTypeApply   TerraformDeployRunType = "apply"
	TerraformDeployRunTypeUnknown TerraformDeployRunType = "unknown"
)

type TerraformDeployS3Archive struct {
	Bucket    string             `hcl:"bucket" validate:"required"`
	BucketKey string             `hcl:"bucket_key" validate:"required"`
	Auth      credentials.Config `hcl:"auth" validate:"required"`
}

type TerraformDeployDirArchive struct {
	Path string `hcl:"path" validate:"required"`
}

type TerraformDeployOCIArchive struct {
	// NOTE(jm): we can not pull the archive information in from the registry plugin, as waypoint doesn't
	// support that.
	//
	// FWIW, we could just share the code here + config between this and the registry, but that probably
	// needs a bit more refactoring as the build + deploy sides are fairly, fairly decoupled.
	Username    string `hcl:"username" validate:"required"`
	AuthToken   string `hcl:"auth_token" validate:"required"`
	RegistryURL string `hcl:"registry_url" validate:"required"`
	Repo        string `hcl:"repo" validate:"required"`
	Tag         string `hcl:"tag" validate:"required"`
}

type TerraformDeployBackend struct {
	Bucket   string             `hcl:"bucket" validate:"required"`
	StateKey string             `hcl:"state_key" validate:"required"`
	Region   string             `hcl:"region" validate:"required"`
	Auth     credentials.Config `hcl:"aws_auth" validate:"required"`
}

type TerraformDeployOutputs struct {
	Bucket         string             `hcl:"bucket" validate:"required"`
	Auth           credentials.Config `hcl:"aws_auth" validate:"required"`
	JobPrefix      string             `hcl:"job_prefix" validate:"required"`
	InstancePrefix string             `hcl:"instance_prefix" validate:"required"`
}

type TerraformDeployHooks struct {
	Enabled bool               `hcl:"enabled"`
	EnvVars map[string]string  `hcl:"env_vars"`
	RunAuth credentials.Config `hcl:"run_auth,block"`
}

type TerraformDeploy struct {
	Plugin string `hcl:"plugin,label"`

	OCIArchive *TerraformDeployOCIArchive `hcl:"oci_archive,block"`
	S3Archive  *TerraformDeployS3Archive  `hcl:"s3_archive,block"`
	DirArchive *TerraformDeployDirArchive `hcl:"local_archive,block"`

	TerraformVersion string                 `hcl:"terraform_version"`
	RunType          TerraformDeployRunType `hcl:"run_type"`

	// auth for the run itself
	RunAuth credentials.Config `hcl:"run_auth,block"`

	// Backend is used to configure where/how the backend is run
	Backend TerraformDeployBackend `hcl:"backend,block"`

	// Outputs are used to control where the run outputs are synchronized to
	Outputs TerraformDeployOutputs `hcl:"outputs,block"`

	Labels    map[string]string `hcl:"labels" validate:"required"`
	Variables map[string]string `hcl:"variables"`
	EnvVars   map[string]string `hcl:"env_vars"`

	// NOTE(jm): I'm not a fan of this approach, but it's faster than building a custom map[string]interface{}
	// decoder for go-hcl. Go-HCL does not support map[string]interface{} out of the box.
	VariablesJSON string                `hcl:"variables_json"`
	Hooks         *TerraformDeployHooks `hcl:"hooks,block"`
}
