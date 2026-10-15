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

// StateRm removes resource addresses from state without destroying the
// underlying infrastructure.
func (w *workspace) StateRm(ctx context.Context, log hclog.Logger, addrs []string) error {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return err
	}

	return w.stateRm(ctx, client, addrs)
}

func (w *workspace) stateRm(ctx context.Context, client Terraform, addrs []string) error {
	for _, addr := range addrs {
		if err := client.StateRm(ctx, addr); err != nil {
			return fmt.Errorf("unable to run state rm %q: %w", addr, err)
		}
	}
	return nil
}
