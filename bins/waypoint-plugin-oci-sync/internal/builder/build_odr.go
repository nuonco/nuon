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
	_ hclog.Logger,
	accessInfo *ociv1.AccessInfo,
) (*ociv1.BuildOutput, error) {
	// create a logger with the output of the ui
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return nil, fmt.Errorf("unable to get output writers: %w", err)
	}

	log := hclog.New(&hclog.LoggerOptions{
		Name:   "waypoint-plugin-oci-sync",
		Output: stdout,
	})

	log.Info("Setting up to sync OCI arifact from vendor repo to customer repo")
	log.Info("received access info", "accessInfo", accessInfo)

	log.Info("Getting vendor repo")
	vendorRepo, err := b.getSrcRepo()
	if err != nil {
		log.Info("unable to get vendor repo")
		return nil, fmt.Errorf("unable to get vendor repo: %w", err)
	}

	log.Info("Getting customer repo")
	customerRepo, err := b.getDstRepo(accessInfo)
	if err != nil {
		log.Info("unable to get customer repo")
		return nil, fmt.Errorf("unable to get customer repo: %w", err)
	}

	log.Info("Copying OCI artifact from vendor repo to customer repo")
	err = b.copy(ctx, vendorRepo, customerRepo, accessInfo.Tag)
	if err != nil {
		log.Info("unable to copy oci artifact from vendor repo to customer repo")
		return nil, fmt.Errorf("unable to copy oci artifact from vendor repo to customer repo: %w", err)
	}

	return &ociv1.BuildOutput{}, nil
}
