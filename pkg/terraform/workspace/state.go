package workspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
)

// StateMv renames a resource address in state without touching the underlying
// infrastructure. Used to migrate legacy sandbox policy keys (`N.yaml`) to
// content-derived keys without forcing create_before_destroy churn.
func (w *workspace) StateMv(ctx context.Context, log hclog.Logger, source, destination string) error {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return err
	}

	if err := client.StateMv(ctx, source, destination); err != nil {
		return fmt.Errorf("unable to run state mv %q -> %q: %w", source, destination, err)
	}
	return nil
}
