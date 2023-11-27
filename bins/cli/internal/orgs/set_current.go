package orgs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
)

func (s *Service) SetCurrent(ctx context.Context, orgID string, cfg *config.Config) {
	cfg.Set("org_id", orgID)
	cfg.WriteConfig()
	fmt.Printf("%s is now the current org\n", orgID)
}
