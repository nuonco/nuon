package shell

import (
	"context"
)

func (s *shell) Init(ctx context.Context, root string) error {
	s.rootDir = root
	return nil
}
