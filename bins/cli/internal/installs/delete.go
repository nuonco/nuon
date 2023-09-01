package installs

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.api.DeleteInstall(ctx, id)
	if err != nil {
		return err
	}

	ui.Line(ctx, "Install %s was deleted", id)
	return nil
}
