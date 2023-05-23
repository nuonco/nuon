package pipeline

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/pipeline/mappers"
)

// Pipeline is a type that is used to execute various commands in succession, with fail/retry logic as well as callbacks
// and others for sharing+persisting state. It's designed to power workflows such as running a terraform run which may
// involve many different steps (and outputs to s3).
//
// This is designed so that types that need to run these types of workflows can decouple the building of the steps +
// logic, from the actual execution of it.
type Pipeline struct {
	v *validator.Validate

	Steps []*Step `validate:"required,gt=1"`

	// TODO(jm): support both a logger, as well as a UI for running these types of steps.
	log    *log.Logger
	ui     terminal.UI
	mapper mappers.Mapper
}

func New(v *validator.Validate) (*Pipeline, error) {
	return &Pipeline{
		v:     v,
		Steps: make([]*Step, 0),

		log:    nil,
		ui:     nil,
		mapper: mappers.NewDefaultMapper(),
	}, nil
}
