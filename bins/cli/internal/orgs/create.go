package orgs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
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
		switch {
		case strings.Contains(err.Error(), "duplicated key"):
			err = errs.UserFacingError(err, fmt.Sprintf("An organization already exists with the name %q", name))
		default:
			err = errors.Wrap(err, "error creating org")
		}
		view.Fail(err)
		return err
	}

	for {
		s.api.SetOrgID(org.ID)
		o, err := s.api.GetOrg(ctx)
		switch {
		case err != nil:
			view.Fail(err)
			return err
		// TODO (sdboyer) need a separate subsystem for statuses
		case o.Status == statusAccessError:
			view.Fail(err)
			return errors.Newf("failed to create org due to access error: %s", o.StatusDescription)
		case o.Status == statusError:
			return errors.Newf("failed to create org: %s", o.StatusDescription)
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
