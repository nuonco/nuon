package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
)

func (b *Builder) BuildODRFunc() interface{} {
	return b.buildODR
}

// build authenticates with both a vendor OCI repo and a customer OCI repo,
// then copies an OCI artifact from the vendor to the customer.
func (b *Builder) buildODR(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	log hclog.Logger,
	customerAccessInfo *ociv1.AccessInfo,
) (*ociv1.BuildOutput, error) {
	// set up
	sg := ui.StepGroup()
	defer sg.Wait()
	step := sg.Add("Setting up to sync OCI arifact from vendor repo to customer repo")
	defer func() {
		if step != nil {
			step.Abort()
		}
	}()
	step.Done()

	step = sg.Add("Getting vendor repo")
	vendorRepo, err := getRepo(b.config.Auth.RegistryURL, b.config.Auth.Username, b.config.Auth.AuthToken)
	if err != nil {
		step.Update(terminal.StatusError, "unable to get vendor repo")
		return nil, fmt.Errorf("unable to get vendor repo: %w", err)
	}
	step.Done()

	step = sg.Add("Getting customer repo")
	customerRepo, err := getRepo(customerAccessInfo.Image, customerAccessInfo.Auth.Username, customerAccessInfo.Auth.Password)
	if err != nil {
		step.Update(terminal.StatusError, "unable to get customer repo")
		return nil, fmt.Errorf("unable to get customer repo: %w", err)
	}
	step.Done()

	step = sg.Add("Copying OCI artifact from vendor repo to customer repo")
	err = b.copy(ctx, vendorRepo, customerRepo)
	if err != nil {
		step.Update(terminal.StatusError, "unable to copy oci artifact from vendor repo to customer repo")
		return nil, fmt.Errorf("unable to copy oci artifact from vendor repo to customer repo: %w", err)
	}
	step.Done()

	return &ociv1.BuildOutput{}, nil
}
