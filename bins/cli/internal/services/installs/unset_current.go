package installs

import (
	"context"
)

func (s *Service) UnsetCurrent(ctx context.Context) error {

	if err := s.unsetInstallID(ctx); err != nil {
		return err
	}

	s.printInstallUnsetMsg()
	return nil
}
