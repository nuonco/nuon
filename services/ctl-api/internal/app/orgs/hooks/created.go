package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (o *Hooks) Created(ctx context.Context, orgID string) {
	o.l.Info("org created", zap.String("id", orgID))
}
