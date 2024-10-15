package helm

import (
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

func (h *handler) packageChart(l *zap.Logger) (string, error) {
	chartDir := h.state.workspace.Source().AbsPath()
	dstDir := h.state.arch.TmpDir()

	chart, err := loader.Load(chartDir)
	if err != nil {
		return "", fmt.Errorf("unable to load chart: %w", err)
	}
	l.Info("succesfully loaded chart")

	packagePath, err := chartutil.Save(chart, dstDir)
	if err != nil {
		return "", fmt.Errorf("unable to package chart: %w", err)
	}
	l.Info("succesfully packaged chart")

	return packagePath, nil
}
