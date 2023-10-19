package shell

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

func (s *shell) PreApply(ctx context.Context, log hclog.Logger) error {
	exists, err := s.existsAndExecutable(hookPreApply)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return s.execScript(ctx, hookPreApply, log)
}
