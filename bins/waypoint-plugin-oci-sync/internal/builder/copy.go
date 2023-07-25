package builder

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
)

// Copies an OCI artifact from a remote vendor repo to a local destination repo.
func (r *Builder) copy(ctx context.Context, vendorRepo oras.ReadOnlyTarget, dstRepo oras.Target, dstTag string) error {
	_, err := oras.Copy(ctx, vendorRepo, r.config.Tag, dstRepo, dstTag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy oci artifact from vendor to customer: %w", err)
	}

	return nil
}
