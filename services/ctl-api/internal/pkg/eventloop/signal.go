package eventloop

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type SignalType string

type Signal interface {
	WorkflowID(id string) string
	WorkflowName() string
	Namespace() string
	Name() string
	SignalType() SignalType

	// for managing context
	GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error)
	PropagateContext(ctx cctx.ValueContext) error
	GetWorkflowContext(ctx workflow.Context) workflow.Context
	GetContext(ctx context.Context) context.Context

	// lifecycle methods
	Validate(*validator.Validate) error
	Stop() bool
	Restart() bool
	Start() bool
}

// SignalHandlerWorkflowID returns the standard ID to use for a signal handler child workflow
func SignalHandlerWorkflowID(sig Signal, req EventLoopRequest) string {
	return fmt.Sprintf("sig-%s-%s", sig.SignalType(), req.ID)
}
