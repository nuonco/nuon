package helm

import (
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

func (h *handler) getSourceFiles() ([]ociarchive.FileRef, error) {
	return []ociarchive.FileRef{
		{
			AbsPath: h.state.packagePath,
			RelPath: defaultChartPackageFilename,
		},
	}, nil
}
