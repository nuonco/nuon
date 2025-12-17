package helm

import (
	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
)

func (h *handler) getSourceFiles() ([]ociarchive.FileRef, error) {
	return []ociarchive.FileRef{
		{
			AbsPath: h.state.packagePath,
			RelPath: defaultChartPackageFilename,
		},
	}, nil
}
