package configs

import (
	"github.com/go-playground/validator/v10"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

// baseBuilder is a builder that exposes a set of options and configs for other builders to use. This is useful as many
// of these same options are used in almost every waypoint config we're building.
type baseBuilder struct {
	v *validator.Validate

	EcrRef    *planv1.ECRRepositoryRef `validate:"required"`
	Metadata  *planv1.Metadata         `validate:"required"`
	Component *componentv1.Component   `validate:"required"`
}

type baseBuilderOption func(*baseBuilder) error

func newBaseBuilder(v *validator.Validate, opts ...baseBuilderOption) (*baseBuilder, error) {
	bld := &baseBuilder{
		v: v,
	}

	for _, opt := range opts {
		if err := opt(bld); err != nil {
			return nil, err
		}
	}

	if err := bld.v.Struct(bld); err != nil {
		return nil, err
	}

	return bld, nil
}

func WithMetadata(metadata *planv1.Metadata) baseBuilderOption {
	return func(b *baseBuilder) error {
		b.Metadata = metadata
		return nil
	}
}

func WithEcrRef(ecrRef *planv1.ECRRepositoryRef) baseBuilderOption {
	return func(b *baseBuilder) error {
		b.EcrRef = ecrRef
		return nil
	}
}

func WithComponent(comp *componentv1.Component) baseBuilderOption {
	return func(b *baseBuilder) error {
		b.Component = comp
		return nil
	}
}
