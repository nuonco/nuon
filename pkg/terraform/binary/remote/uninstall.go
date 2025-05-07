package remote

import (
	"context"
)

func (r *remote) Uninstall(ctx context.Context) error {
	if r.version != nil {
		return r.version.Remove(ctx)
	}

	return nil
}
