package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateStaticToken(ctx context.Context, name, duration, role string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if name == "" {
		return view.Error(fmt.Errorf("name is required"))
	}

	token, err := s.api.CreateStaticToken(ctx, &models.ServiceCreateStaticTokenRequest{
		Name:     &name,
		Duration: &duration,
		Role:     role,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(token)
		return nil
	}

	view.Render([][]string{
		{"id", token.ID},
		{"api token", token.APIToken},
	})
	fmt.Println("\nCopy this token now. For security, it won't be shown again.")
	return nil
}
