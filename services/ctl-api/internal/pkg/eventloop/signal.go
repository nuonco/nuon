package eventloop

import (
	"context"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SignalType string

type Signal interface {
	WorkflowID(id string) string
	WorkflowName() string
	Namespace() string
	Name() string
	SignalType() SignalType
	GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error)

	// used to control different behaviours on the event loop
	// Fail the workflow when an error is returned
	FailOnError() bool

	// Stop the workflow when finished
	StopOnFinish() bool

	// Start the workflow, or restart it
	Start() bool

	// DoNothing
	Noop() bool
}
