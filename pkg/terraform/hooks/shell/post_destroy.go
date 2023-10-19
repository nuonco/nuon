package shell

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

func (s *shell) PostDestroy(ctx context.Context, log hclog.Logger) error {
	exists, err := s.existsAndExecutable(hookPostDestroy)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return s.execScript(ctx, hookPostDestroy, log)
}
