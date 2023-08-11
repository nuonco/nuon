package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (o *hooks) AfterCreate(ctx context.Context, id string) {
	o.l.Info("hello world in after create", zap.String("id", id))
}
