package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

func (h *handler) packageChart() (string, error) {
	chartDir := h.state.workspace.Source().Path
	dstDir := h.state.arch.TmpDir()

	chart, err := loader.Load(chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to load chart: %w", err)
	}
	h.log.Info("succesfully loaded chart")

	packagePath, err := chartutil.Save(chart, dstDir)
	if err != nil {
		return "", fmt.Errorf("unable to package chart: %w", err)
	}
	h.log.Info("succesfully packaged chart")

	return packagePath, nil
}
