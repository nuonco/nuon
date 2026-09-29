package workspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
)

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
