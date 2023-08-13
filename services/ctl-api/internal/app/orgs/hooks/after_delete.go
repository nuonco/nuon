package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (o *Hooks) Deleted(ctx context.Context, orgID string) {
	o.l.Info("org deleted", zap.String("org", orgID))
}
