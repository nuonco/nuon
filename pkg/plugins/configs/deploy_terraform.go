package configs

import "github.com/powertoolsdev/mono/pkg/aws/credentials"

type TerraformDeploy struct {
	Name string `hcl:"name,label"`

	Archive struct {
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
	} `hcl:"archive,block"`

	TerraformVersion string `hcl:"terraform_version"`

	// auth for the run itself
	RunAuth  credentials.Config `hcl:"run_auth,block"`
	PlanOnly bool               `hcl:"plan_only"`

	// outputs are used to set the outputs after the terraform run
	Backend struct {
		Bucket   string             `hcl:"bucket" validate:"required"`
		StateKey string             `hcl:"state_key" validate:"required"`
		Region   string             `hcl:"region" validate:"required"`
		Auth     credentials.Config `hcl:"aws_auth" validate:"required"`
	} `hcl:"backend,block"`

	// Outputs are used to control where the run outputs are synchronized to
	Outputs struct {
		Bucket string             `hcl:"bucket" validate:"required"`
		Auth   credentials.Config `hcl:"aws_auth" validate:"required"`
		Prefix string             `hcl:"prefix" validate:"required"`
	} `hcl:"outputs,block"`

	Labels    map[string]string `hcl:"labels" validate:"required"`
	Variables map[string]string `hcl:"variables" validate:"required"`
}
