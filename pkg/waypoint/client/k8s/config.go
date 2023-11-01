package k8s

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type Token struct {
	Namespace, Name, Key string `validate:"required"`
}

type Config struct {
	Address     string `validate:"required"`
	ClusterInfo *kube.ClusterInfo
	Token       Token `validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("unable to validate config: %w", err)
	}

	return nil
}
