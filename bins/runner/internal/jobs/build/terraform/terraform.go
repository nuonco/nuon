package terraform

import (
	"context"

	"github.com/go-playground/validator/v10"
	"oras.land/oras-go/v2/content/file"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

type builder struct {
	v *validator.Validate

	config configs.TerraformBuildAWSECRRegistry
	Store  *file.Store
}

func (b *builder) SetConfig(map[string]interface{}) error {
	return nil
}

func (b *builder) Initialize(ctx context.Context) error {
	return nil
}

func (b *builder) Build(ctx context.Context) error {
	return nil
}

func (b *builder) Push(ctx context.Context) error {
	return nil
}

func (b *builder) Cleanup(ctx context.Context) error {
	return nil
}
