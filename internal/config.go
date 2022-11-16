package workers

import "github.com/go-playground/validator/v10"

type Config struct {
	OrgsIamRoleArn string `config:"orgs_account_iam_role_arn" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
