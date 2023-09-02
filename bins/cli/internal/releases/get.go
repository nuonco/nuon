package releases

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Get(ctx context.Context, releaseID string) error {
	release, err := s.api.GetRelease(ctx, releaseID)
	if err != nil {
		return err
	}
	ui.Line(ctx, "%s - %s", release.ID, release.Status)

	return nil
}
