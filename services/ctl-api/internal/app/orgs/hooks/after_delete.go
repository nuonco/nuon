package hooks

import "context"

func (o *hooks) AfterDelete(ctx context.Context, orgID string) {
	o.l.Info("after delete")
}
