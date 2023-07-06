package builder

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
)

// Copies an OCI artifact from a remote vendor repo to a local customer repo.
func (r *Builder) copy(ctx context.Context, vendorRepo oras.ReadOnlyTarget, customerRepo oras.Target) error {
	_, err := oras.Copy(ctx, vendorRepo, r.config.Tag, customerRepo, r.config.Tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy oci artifact from vendor to customer: %w", err)
	}

	return nil
}
