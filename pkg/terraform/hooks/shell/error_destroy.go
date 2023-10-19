package shell

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

func (s *shell) ErrorDestroy(ctx context.Context, log hclog.Logger) error {
	exists, err := s.existsAndExecutable(hookErrorDestroy)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return s.execScript(ctx, hookErrorDestroy, log)
}
