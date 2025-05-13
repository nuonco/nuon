package eventloop

import (
	"context"
	"fmt"
	"strings"

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

	// for managing intra-workflow communication
	Listeners() []SignalListener

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
func SignalHandlerWorkflowID(ctx workflow.Context, sig Signal, req EventLoopRequest) string {
	// If the calling workflow is an event loop, then construct this name hierarchically
	caller := workflow.GetInfo(ctx).WorkflowExecution.ID
	if strings.HasPrefix(caller, "event-loop-") {
		if len(caller) > len(req.ID) && strings.HasSuffix(caller, req.ID) {
			// If the ID is the same as the caller's, don't repeat it
			return fmt.Sprintf("%s::%s", caller, sig.SignalType())
		} else {
			return fmt.Sprintf("%s::%s-%s", caller, sig.SignalType(), req.ID)
		}
	}
	return fmt.Sprintf("sig-%s-%s", sig.SignalType(), req.ID)
}
