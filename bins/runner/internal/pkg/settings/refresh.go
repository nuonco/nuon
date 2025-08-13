package settings

import (
	"context"
)

func (s *Settings) Refresh(ctx context.Context) error {
	err := s.fetch(ctx)
	return err
}
