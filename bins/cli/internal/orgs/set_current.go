package orgs

import (
	"context"
	"fmt"
)

func (s *Service) SetCurrent(ctx context.Context, orgID string, showMsg bool) {
	s.cfg.Set("org_id", orgID)
	s.cfg.WriteConfig()
	if showMsg == true {
		fmt.Printf("%s is now the current org\n", orgID)
	}
}
