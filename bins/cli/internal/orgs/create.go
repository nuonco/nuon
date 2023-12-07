package orgs

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access-error"
)

func (s *Service) Create(ctx context.Context, name string, isSandboxMode bool, asJSON bool) {
	if asJSON {
		org, err := s.api.CreateOrg(ctx, &models.ServiceCreateOrgRequest{
			Name:           &name,
			UseSandboxMode: isSandboxMode,
		})
		if err != nil {
			ui.PrintJSONError(err)
			return
		}
		ui.PrintJSON(org)
		s.SetCurrent(ctx, org.ID, false)
		return
	}

	view := ui.NewCreateView("org", asJSON)
	view.Start()
	view.Update("creating org")
	org, err := s.api.CreateOrg(ctx, &models.ServiceCreateOrgRequest{
		Name:           &name,
		UseSandboxMode: isSandboxMode,
	})
	if err != nil {
		view.Fail(err)
		return
	}

	for {
		s.api.SetOrgID(org.ID)
		o, err := s.api.GetOrg(ctx)
		switch {
		case err != nil:
			view.Fail(err)
		case o.Status == statusAccessError:
			view.Fail(fmt.Errorf("failed to create org due to access error: %s", o.StatusDescription))
			return
		case o.Status == statusError:
			view.Fail(fmt.Errorf("failed to create org: %s", o.StatusDescription))
			return
		case o.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created org %s", o.ID))
			s.SetCurrent(ctx, o.ID, false)
			return
		default:
			view.Update(fmt.Sprintf("%s org", o.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
