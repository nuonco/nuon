package orgs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/errs"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access-error"
)

func (s *Service) Create(ctx context.Context, name string, isSandboxMode bool, asJSON bool) error {
	if asJSON {
		org, err := s.api.CreateOrg(ctx, &models.ServiceCreateOrgRequest{
			Name:           &name,
			UseSandboxMode: isSandboxMode,
		})
		if err != nil {
			ui.PrintJSONError(err)
			return err
		}
		ui.PrintJSON(org)
		s.setOrgInConfig(ctx, org.ID)
		return err
	}

	view := ui.NewCreateView("org", asJSON)
	view.Start()
	view.Update("creating org")
	org, err := s.api.CreateOrg(ctx, &models.ServiceCreateOrgRequest{
		Name:           &name,
		UseSandboxMode: isSandboxMode,
	})
	if err != nil {
		// TODO(sdboyer) this kind of string sniffing will be replaced when deep leaf errors are managed by the system
		if strings.Contains(err.Error(), "duplicated key") {
			err = errs.WithUserFacing(err, fmt.Sprintf("An organization already exists with the name %q", name))
		}
		return view.Fail(err)
	}

	for {
		s.api.SetOrgID(org.ID)
		o, err := s.api.GetOrg(ctx)
		switch {
		case err != nil:
			return view.Fail(err)
		// TODO (sdboyer) need a separate subsystem for statuses
		case o.Status == statusAccessError:
			return view.Fail(errs.NewUserFacing("failed to create org due to access error: %s", o.StatusDescription))
		case o.Status == statusError:
			return view.Fail(errs.NewUserFacing("failed to create org: %s", o.StatusDescription))
		case o.Status == statusActive:
			view.Success(fmt.Sprintf("successfully created org %s", o.ID))
			s.setOrgInConfig(ctx, o.ID)
			return nil
		default:
			view.Update(fmt.Sprintf("%s org", o.Status))
		}

		time.Sleep(5 * time.Second)
	}
}
