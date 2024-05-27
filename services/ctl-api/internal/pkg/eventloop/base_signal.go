package eventloop

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type BaseSignal struct{}

func (BaseSignal) WorkflowName() string {
	return "EventLoop"
}

func (BaseSignal) WorkflowID(id string) string {
	return "event-loop-" + id
}

func (BaseSignal) FailOnError() bool {
	return false
}

func (BaseSignal) StopOnFinish() bool {
	return false
}

func (BaseSignal) Noop() bool {
	return false
}

func (BaseSignal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := org.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get org from context: %w", err)
	}

	return org, nil
}
