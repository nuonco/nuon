package shell

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

func (s *shell) PreDestroy(ctx context.Context, log hclog.Logger) error {
	exists, err := s.existsAndExecutable(hookPreDestroy)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return s.execScript(ctx, hookPreDestroy, log)
}
