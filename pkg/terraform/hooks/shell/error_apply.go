package shell

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

func (s *shell) ErrorApply(ctx context.Context, log hclog.Logger) error {
	exists, err := s.existsAndExecutable(hookErrorApply)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return s.execScript(ctx, hookErrorApply, log)
}
