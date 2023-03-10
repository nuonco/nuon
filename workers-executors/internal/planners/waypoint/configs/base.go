package configs

import (
	"github.com/go-playground/validator/v10"
	buildv1 "github.com/powertoolsdev/protos/components/generated/types/build/v1"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
)

// baseBuilder is a builder that exposes a set of options and configs for other builders to use. This is useful as many
// of these same options are used in almost every waypoint config we're building.
type baseBuilder struct {
	v *validator.Validate

	EcrRef      *planv1.ECRRepositoryRef `validate:"required"`
	WaypointRef *planv1.WaypointRef      `validate:"required"`
	Component   *componentv1.Component   `validate:"required"`

	// optional params, for different config builders. By default, we try to use the component protos where
	// possible.
	PrivateImageSource *PrivateImageSource
	PublicImageSource  *PublicImageSource
	DockerCfg          *buildv1.DockerConfig
}

type Option func(*baseBuilder) error

func newBaseBuilder(v *validator.Validate, opts ...Option) (*baseBuilder, error) {
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

// WithEcrRef is used to pass in a configuration for pushing an image to ecr
func WithEcrRef(ecrRef *planv1.ECRRepositoryRef) Option {
	return func(b *baseBuilder) error {
		b.EcrRef = ecrRef
		return nil
	}
}

func WithComponent(comp *componentv1.Component) Option {
	return func(b *baseBuilder) error {
		// TODO(jm): add validating
		//if err := comp.ValidateAll(); err != nil {
		//return err
		//}

		b.Component = comp
		return nil
	}
}

func WithWaypointRef(ref *planv1.WaypointRef) Option {
	return func(b *baseBuilder) error {
		b.WaypointRef = ref
		return nil
	}
}

func WithPrivateImageSource(img *PrivateImageSource) Option {
	return func(b *baseBuilder) error {
		if err := img.validate(b.v); err != nil {
			return err
		}

		b.PrivateImageSource = img
		return nil
	}
}

func WithPublicImageSource(img *PublicImageSource) Option {
	return func(b *baseBuilder) error {
		if err := img.validate(b.v); err != nil {
			return err
		}

		b.PublicImageSource = img
		return nil
	}
}

func WithDockerCfg(cfg *buildv1.DockerConfig) Option {
	return func(b *baseBuilder) error {
		b.DockerCfg = cfg
		return nil
	}
}
