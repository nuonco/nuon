package opentofu

import (
	"bytes"
	"context"

	"github.com/pkg/errors"

	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

const (
	// backendBlock is the same s3-backend stub the terraform build validates
	// for. OpenTofu parses the same HCL block, and the ctl-api backend file
	// injection writes the same shape for both engines, so the check is
	// identical.
	backendBlock string = "backend s3 {}"
)

func (h *handler) validateSourceFiles(ctx context.Context, source *sourcefs.FS, files []ociarchive.FileRef) error {
	for _, src := range files {
		byts, err := source.ReadFile(src.AbsPath)
		if err != nil {
			return errors.Wrap(err, "unable to read file")
		}

		if bytes.Contains(byts, []byte(backendBlock)) {
			return nil
		}
	}

	return errors.New("no backend s3 {} block found")
}
