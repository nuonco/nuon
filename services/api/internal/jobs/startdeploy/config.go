package startdeploy

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {

	// bucket configurations
	DeploymentsBucket string `config:"deployments_bucket" validate:"required"`
	// org IAM role template names
	OrgsDeploymentsRoleTemplate string `config:"orgs_deployments_role_template" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
